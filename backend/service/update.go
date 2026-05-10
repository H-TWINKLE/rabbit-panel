package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"rabbit-panel/repository"
)

const (
	updateManifestURL = "https://reisen7.github.io/rabbit-panel/latest.json"
	updateLogPath     = "/tmp/rabbit-panel-update.log"
	updateScriptPath  = "/tmp/rabbit-panel-update.sh"
	updateBinaryDir   = "/tmp/rabbit-panel-update"
)

type BuildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

type DeployConfig struct {
	Mode           string `json:"mode"`
	Image          string `json:"image"`
	ImageTag       string `json:"image_tag"`
	HostProjectDir string `json:"host_project_dir"`
	ComposeFile    string `json:"compose_file"`
	ServiceName    string `json:"service_name"`
}

type UpdateManifest struct {
	Version      string            `json:"version"`
	VersionNoV   string            `json:"version_number"`
	ReleaseURL   string            `json:"release_url"`
	ReleaseName  string            `json:"release_name"`
	ReleaseNotes string            `json:"release_notes"`
	PublishedAt  string            `json:"published_at"`
	DockerImage  string            `json:"docker_image"`
	DockerTag    string            `json:"docker_tag"`
	DockerTags   []string          `json:"docker_tags"`
	Assets       map[string]string `json:"assets"`
	SHA256       map[string]string `json:"sha256"`
}

type UpdateCheckResult struct {
	CurrentVersion   string `json:"current_version"`
	CurrentCommit    string `json:"current_commit"`
	CurrentBuildTime string `json:"current_build_time"`
	LatestVersion    string `json:"latest_version"`
	HasUpdate        bool   `json:"has_update"`
	DeployMode       string `json:"deploy_mode"`
	Image            string `json:"image"`
	ImageTag         string `json:"image_tag"`
	CanUpdate        bool   `json:"can_update"`
	IgnoredVersion   string `json:"ignored_version"`
	Ignored          bool   `json:"ignored"`
	ReleaseURL       string `json:"release_url"`
	ReleaseNotes     string `json:"release_notes"`
	Message          string `json:"message"`
	LastCheckTime    string `json:"last_check_time"`
	LastUpdateTime   string `json:"last_update_time"`
	LastUpdateStatus string `json:"last_update_status"`
	LastUpdateError  string `json:"last_update_error"`
}

type UpdateTaskStatus struct {
	Status         string   `json:"status"`
	Stage          string   `json:"stage"`
	Progress       int      `json:"progress"`
	ProgressKnown  bool     `json:"progress_known"`
	LastUpdateTime string `json:"last_update_time"`
	LastError      string   `json:"last_error"`
	LogLines       []string `json:"log_lines"`
}

type UpdateSettings struct {
	IgnoredVersion   string `json:"ignored_version"`
	LastCheckTime    string `json:"last_check_time"`
	LastUpdateTime   string `json:"last_update_time"`
	LastUpdateStatus string `json:"last_update_status"`
	LastUpdateError  string `json:"last_update_error"`
}

type UpdateService struct {
	fileRepo   repository.IFileRepository
	buildInfo  BuildInfo
	httpClient *http.Client
	mu         sync.Mutex
}

func NewUpdateService(fileRepo repository.IFileRepository, buildInfo BuildInfo) *UpdateService {
	return &UpdateService{
		fileRepo:  fileRepo,
		buildInfo: buildInfo,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *UpdateService) GetBuildInfo() BuildInfo {
	return s.buildInfo
}

func (s *UpdateService) DetectDeployConfig() DeployConfig {
	mode := strings.TrimSpace(os.Getenv("RABBIT_DEPLOY_MODE"))
	if mode == "" {
		mode = detectContainerMode()
	}
	if mode != "docker" && mode != "binary" {
		mode = "binary"
	}

	return DeployConfig{
		Mode:           mode,
		Image:          getEnv("RABBIT_IMAGE", "reisen7/rabbit-panel"),
		ImageTag:       getEnv("RABBIT_IMAGE_TAG", "latest"),
		HostProjectDir: getEnv("RABBIT_HOST_PROJECT_DIR", "/root/rabbit-panel"),
		ComposeFile:    getEnv("RABBIT_COMPOSE_FILE", "docker-compose.deploy.yml"),
		ServiceName:    getEnv("RABBIT_SERVICE_NAME", "rabbit-panel"),
	}
}

func (s *UpdateService) Check(ctx context.Context) (*UpdateCheckResult, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return nil, err
	}

	manifest, err := s.fetchManifest(ctx)
	if err != nil {
		return nil, err
	}

	deploy := s.DetectDeployConfig()
	currentVersion := normalizeVersion(s.buildInfo.Version)
	latestVersion := normalizeVersion(manifest.Version)

	hasUpdate := false
	if currentVersion != "" && currentVersion != "dev" && latestVersion != "" {
		cmp, cmpErr := compareVersions(currentVersion, latestVersion)
		hasUpdate = cmpErr == nil && cmp < 0
	}

	ignored := settings.IgnoredVersion != "" && normalizeVersion(settings.IgnoredVersion) == latestVersion
	if ignored || currentVersion == "dev" {
		hasUpdate = false
	}

	settings.LastCheckTime = time.Now().Format(time.RFC3339)
	if err := s.saveSettings(settings); err != nil {
		log.Printf("save update settings failed: %v", err)
	}

	return &UpdateCheckResult{
		CurrentVersion:   s.buildInfo.Version,
		CurrentCommit:    s.buildInfo.Commit,
		CurrentBuildTime: s.buildInfo.BuildTime,
		LatestVersion:    manifest.Version,
		HasUpdate:        hasUpdate,
		DeployMode:       deploy.Mode,
		Image:            deploy.Image,
		ImageTag:         deploy.ImageTag,
		CanUpdate:        canUpdate(deploy),
		IgnoredVersion:   settings.IgnoredVersion,
		Ignored:          ignored,
		ReleaseURL:       manifest.ReleaseURL,
		ReleaseNotes:     manifest.ReleaseNotes,
		Message:          buildUpdateMessage(s.buildInfo.Version, manifest.Version, hasUpdate, ignored),
		LastCheckTime:    settings.LastCheckTime,
		LastUpdateTime:   settings.LastUpdateTime,
		LastUpdateStatus: settings.LastUpdateStatus,
		LastUpdateError:  settings.LastUpdateError,
	}, nil
}

func (s *UpdateService) IgnoreVersion(version string) error {
	settings, err := s.loadSettings()
	if err != nil {
		return err
	}
	settings.IgnoredVersion = strings.TrimSpace(version)
	return s.saveSettings(settings)
}

func (s *UpdateService) ClearIgnoredVersion() error {
	settings, err := s.loadSettings()
	if err != nil {
		return err
	}
	settings.IgnoredVersion = ""
	return s.saveSettings(settings)
}

func (s *UpdateService) ClearUpdateState() error {
	settings, err := s.loadSettings()
	if err != nil {
		return err
	}
	settings.LastUpdateTime = ""
	settings.LastUpdateStatus = ""
	settings.LastUpdateError = ""
	return s.saveSettings(settings)
}

func (s *UpdateService) StartUpdate(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	deploy := s.DetectDeployConfig()
	if deploy.Mode == "docker" && !canUpdate(deploy) {
		return errors.New("current Docker deployment uses a fixed image tag; switch to latest or update manually")
	}
	if deploy.Mode != "docker" && deploy.Mode != "binary" {
		return errors.New("unsupported deploy mode for update")
	}
	if deploy.Mode == "binary" {
		if err := ensureSystemdServiceExists(deploy.ServiceName); err != nil {
			return err
		}
	}

	manifest, err := s.fetchManifest(ctx)
	if err != nil {
		return err
	}

	if deploy.Mode == "binary" {
		return s.startBinaryUpdate(manifest, deploy)
	}
	return s.startDockerUpdate(deploy)
}

func (s *UpdateService) GetTaskStatus() (*UpdateTaskStatus, error) {
	settings, err := s.loadSettings()
	if err != nil {
		return nil, err
	}

	logLines := readLastLogLines(updateLogPath, 80)
	stage, progress, progressKnown := deriveStage(logLines, settings.LastUpdateStatus)

	return &UpdateTaskStatus{
		Status:         settings.LastUpdateStatus,
		Stage:          stage,
		Progress:       progress,
		ProgressKnown:  progressKnown,
		LastUpdateTime: settings.LastUpdateTime,
		LastError:      settings.LastUpdateError,
		LogLines:       logLines,
	}, nil
}

func (s *UpdateService) loadSettings() (*UpdateSettings, error) {
	record, err := s.fileRepo.LoadUpdateSettings()
	if err != nil {
		return nil, err
	}
	return &UpdateSettings{
		IgnoredVersion:   record.IgnoredVersion,
		LastCheckTime:    record.LastCheckTime,
		LastUpdateTime:   record.LastUpdateTime,
		LastUpdateStatus: record.LastUpdateStatus,
		LastUpdateError:  record.LastUpdateError,
	}, nil
}

func (s *UpdateService) saveSettings(settings *UpdateSettings) error {
	return s.fileRepo.SaveUpdateSettings(&repository.UpdateSettingsRecord{
		IgnoredVersion:   settings.IgnoredVersion,
		LastCheckTime:    settings.LastCheckTime,
		LastUpdateTime:   settings.LastUpdateTime,
		LastUpdateStatus: settings.LastUpdateStatus,
		LastUpdateError:  settings.LastUpdateError,
	})
}

func (s *UpdateService) updateStatus(status, errMsg string) {
	settings, err := s.loadSettings()
	if err != nil {
		log.Printf("load update settings failed: %v", err)
		return
	}
	settings.LastUpdateStatus = status
	settings.LastUpdateError = errMsg
	settings.LastUpdateTime = time.Now().Format(time.RFC3339)
	if err := s.saveSettings(settings); err != nil {
		log.Printf("save update status failed: %v", err)
	}
}

func (s *UpdateService) fetchManifest(ctx context.Context) (*UpdateManifest, error) {
	manifestURL := getEnv("RABBIT_UPDATE_MANIFEST_URL", updateManifestURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, manifestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "rabbit-panel")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("update manifest request failed from %s: %s %s", manifestURL, resp.Status, strings.TrimSpace(string(body)))
	}

	var manifest UpdateManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}
	if manifest.Version == "" {
		return nil, errors.New("update manifest missing version")
	}
	return &manifest, nil
}

func (s *UpdateService) startBinaryUpdate(manifest *UpdateManifest, deploy DeployConfig) error {
	assetURL, assetName, archKey, err := findBinaryAsset(manifest.Assets)
	if err != nil {
		return err
	}
	expectedSHA := strings.TrimSpace(manifest.SHA256[archKey])

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(updateBinaryDir, 0755); err != nil {
		return err
	}

script := fmt.Sprintf(`#!/bin/sh
set -eu
LOG_FILE=%s
mkdir -p "$(dirname "$LOG_FILE")"
: >"$LOG_FILE"
exec >>"$LOG_FILE" 2>&1
echo "[%s] binary update started"
TMP_DIR=%s
mkdir -p "$TMP_DIR"
NEW_BIN="$TMP_DIR/%s"
CURRENT_BIN=%s
BACKUP_BIN="${CURRENT_BIN}.bak"
DOWNLOAD_URL=%s
TOTAL_SIZE=0
if command -v curl >/dev/null 2>&1; then
  TOTAL_SIZE=$(curl -fsSLI "$DOWNLOAD_URL" | awk 'BEGIN{IGNORECASE=1} /^Content-Length:/ {gsub("\r","",$2); print $2}' | tail -n 1 || true)
fi
if [ -z "$TOTAL_SIZE" ] && command -v wget >/dev/null 2>&1; then
  TOTAL_SIZE=$(wget --server-response --spider "$DOWNLOAD_URL" 2>&1 | awk 'BEGIN{IGNORECASE=1} /^  Content-Length:/ {print $2}' | tail -n 1 || true)
fi
if [ -z "$TOTAL_SIZE" ]; then TOTAL_SIZE=0; fi
download_with_progress() {
  DOWNLOADER_PID=$1
  LAST_PERCENT=-1
  while kill -0 "$DOWNLOADER_PID" 2>/dev/null; do
    if [ -f "$NEW_BIN" ] && [ "$TOTAL_SIZE" -gt 0 ] 2>/dev/null; then
      CURRENT_SIZE=$(stat -c%%s "$NEW_BIN" 2>/dev/null || echo 0)
      if [ "$CURRENT_SIZE" -gt 0 ] 2>/dev/null; then
        PERCENT=$((CURRENT_SIZE * 100 / TOTAL_SIZE))
        if [ "$PERCENT" -gt 69 ]; then PERCENT=69; fi
        if [ "$PERCENT" -ne "$LAST_PERCENT" ]; then
          echo "PROGRESS:$PERCENT"
          LAST_PERCENT=$PERCENT
        fi
      fi
    fi
    sleep 1
  done
}
if command -v curl >/dev/null 2>&1; then
  if curl --version 2>/dev/null | grep -qiE 'https|openssl|mbedtls|gnutls|securetransport|schannel'; then
    curl -L --fail --retry 3 -sS -o "$NEW_BIN" "$DOWNLOAD_URL" &
    DL_PID=$!
    download_with_progress "$DL_PID"
    wait "$DL_PID"
  else
    echo "curl found but https is not supported, fallback to wget"
    if command -v wget >/dev/null 2>&1; then
      wget -q -O "$NEW_BIN" "$DOWNLOAD_URL" &
      DL_PID=$!
      download_with_progress "$DL_PID"
      wait "$DL_PID"
    else
      echo "no https-capable downloader available"
      exit 1
    fi
  fi
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$NEW_BIN" "$DOWNLOAD_URL" &
  DL_PID=$!
  download_with_progress "$DL_PID"
  wait "$DL_PID"
else
  echo "neither curl nor wget is available"
  exit 1
fi
echo "PROGRESS:69"
chmod +x "$NEW_BIN"
EXPECTED_SHA=%s
if [ -n "$EXPECTED_SHA" ]; then
  if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL_SHA=$(sha256sum "$NEW_BIN" | awk '{print $1}')
  elif command -v shasum >/dev/null 2>&1; then
    ACTUAL_SHA=$(shasum -a 256 "$NEW_BIN" | awk '{print $1}')
  elif command -v openssl >/dev/null 2>&1; then
    ACTUAL_SHA=$(openssl dgst -sha256 "$NEW_BIN" | awk '{print $NF}')
  else
    echo "no sha256 tool available"
    exit 1
  fi
  if [ "$ACTUAL_SHA" != "$EXPECTED_SHA" ]; then
    echo "sha256 mismatch: expected $EXPECTED_SHA got $ACTUAL_SHA"
    exit 1
  fi
  echo "sha256 verified"
fi
if [ ! -f "$CURRENT_BIN" ]; then
  echo "current binary not found: $CURRENT_BIN"
  exit 1
fi
cp "$CURRENT_BIN" "$BACKUP_BIN"
systemctl stop %s
cp "$NEW_BIN" "$CURRENT_BIN"
chmod +x "$CURRENT_BIN"
if ! systemctl start %s; then
  cp "$BACKUP_BIN" "$CURRENT_BIN"
  systemctl start %s || true
  exit 1
fi
`, shellQuote(updateLogPath), time.Now().Format(time.RFC3339), shellQuote(updateBinaryDir), shellQuote(assetName), shellQuote(execPath), shellQuote(assetURL), shellQuote(expectedSHA), shellQuote(deploy.ServiceName), shellQuote(deploy.ServiceName), shellQuote(deploy.ServiceName), shellQuote(deploy.ServiceName))

	if err := os.WriteFile(updateScriptPath, []byte(script), 0700); err != nil {
		return err
	}

	s.updateStatus("running", "")
	cmd := exec.Command("nohup", "sh", updateScriptPath)
	if err := cmd.Start(); err != nil {
		s.updateStatus("failed", err.Error())
		return err
	}
	go s.watchProcess(cmd)
	return nil
}

func (s *UpdateService) startDockerUpdate(deploy DeployConfig) error {
script := fmt.Sprintf(`#!/bin/sh
set -eu
LOG_FILE=%s
mkdir -p "$(dirname "$LOG_FILE")"
: >"$LOG_FILE"
exec >>"$LOG_FILE" 2>&1
echo "[%s] docker update started"
PROJECT_DIR=%s
COMPOSE_FILE=%s
if command -v docker >/dev/null 2>&1; then
  cd "$PROJECT_DIR"
  docker compose -f "$COMPOSE_FILE" pull
  docker compose -f "$COMPOSE_FILE" up -d
else
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -v "$PROJECT_DIR":"$PROJECT_DIR" -w "$PROJECT_DIR" docker:cli sh -c "docker compose -f \"$COMPOSE_FILE\" pull && docker compose -f \"$COMPOSE_FILE\" up -d"
fi
`, shellQuote(updateLogPath), time.Now().Format(time.RFC3339), shellQuote(deploy.HostProjectDir), shellQuote(deploy.ComposeFile))

	if err := os.WriteFile(updateScriptPath, []byte(script), 0700); err != nil {
		return err
	}

	s.updateStatus("running", "")
	cmd := exec.Command("nohup", "sh", updateScriptPath)
	if err := cmd.Start(); err != nil {
		s.updateStatus("failed", err.Error())
		return err
	}
	go s.watchProcess(cmd)
	return nil
}

func (s *UpdateService) watchProcess(cmd *exec.Cmd) {
	if err := cmd.Wait(); err != nil {
		s.updateStatus("failed", err.Error())
		return
	}
	s.updateStatus("success", "")
}

func detectContainerMode() string {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "docker"
	}
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := strings.ToLower(string(data))
		for _, keyword := range []string{"docker", "containerd", "kubepods", "podman", "lxc"} {
			if strings.Contains(content, keyword) {
				return "docker"
			}
		}
	}
	return "binary"
}

func canUpdate(deploy DeployConfig) bool {
	if deploy.Mode == "binary" {
		return true
	}
	if deploy.Mode == "docker" && strings.EqualFold(strings.TrimSpace(deploy.ImageTag), "latest") {
		return true
	}
	return false
}

func buildUpdateMessage(currentVersion, latestVersion string, hasUpdate, ignored bool) string {
	if normalizeVersion(currentVersion) == "dev" {
		return fmt.Sprintf("Current version is dev, latest release is %s", latestVersion)
	}
	if ignored {
		return fmt.Sprintf("Latest version %s is ignored", latestVersion)
	}
	if hasUpdate {
		return fmt.Sprintf("Found new version %s", latestVersion)
	}
	return "Already on the latest version"
}

func normalizeVersion(version string) string {
	version = strings.TrimSpace(version)
	return strings.TrimPrefix(version, "v")
}

func compareVersions(current, latest string) (int, error) {
	curParts, err := parseVersion(normalizeVersion(current))
	if err != nil {
		return 0, err
	}
	latestParts, err := parseVersion(normalizeVersion(latest))
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if curParts[i] < latestParts[i] {
			return -1, nil
		}
		if curParts[i] > latestParts[i] {
			return 1, nil
		}
	}
	return 0, nil
}

func parseVersion(version string) ([3]int, error) {
	var result [3]int
	parts := strings.Split(version, ".")
	for i := 0; i < len(parts) && i < 3; i++ {
		part := parts[i]
		for idx, ch := range part {
			if ch < '0' || ch > '9' {
				part = part[:idx]
				break
			}
		}
		if part == "" {
			return result, fmt.Errorf("invalid version: %s", version)
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return result, err
		}
		result[i] = value
	}
	return result, nil
}

func findBinaryAsset(assets map[string]string) (string, string, string, error) {
	arch, err := currentReleaseArch()
	if err != nil {
		return "", "", "", err
	}
	downloadURL, ok := assets[arch]
	if !ok || strings.TrimSpace(downloadURL) == "" {
		return "", "", "", fmt.Errorf("no release asset found for %s", arch)
	}
	return downloadURL, filepath.Base(downloadURL), arch, nil
}

func currentReleaseArch() (string, error) {
	switch runtime.GOARCH {
	case "amd64":
		return "linux-amd64", nil
	case "arm64":
		return "linux-arm64", nil
	case "arm":
		return "linux-armv7", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func ensureSystemdServiceExists(serviceName string) error {
	if serviceName == "" {
		serviceName = "rabbit-panel"
	}
	cmd := exec.Command("systemctl", "status", serviceName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("systemd service %s not found; binary auto update requires running Rabbit Panel as a systemd service", serviceName)
	}
	return nil
}

func readLastLogLines(path string, limit int) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return []string{}
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			filtered = append(filtered, line)
		}
	}
	if len(filtered) <= limit {
		return filtered
	}
	return filtered[len(filtered)-limit:]
}

func deriveStage(lines []string, status string) (string, int, bool) {
	lines = extractCurrentRunLines(lines)
	stage := "pending"
	progress := 0
	progressKnown := false
	joined := strings.ToLower(strings.Join(lines, "\n"))
	if p, ok := extractProgress(lines); ok {
		progress = p
		progressKnown = true
		stage = "downloading"
	}

	switch {
	case strings.Contains(joined, "binary update started") || strings.Contains(joined, "docker update started"):
		stage = "started"
	}
	switch {
	case strings.Contains(joined, "wget -o") || strings.Contains(joined, "curl ") || strings.Contains(joined, "fallback to wget"):
		stage = "downloading"
	}
	switch {
	case strings.Contains(joined, "sha256 verified"):
		stage, progress, progressKnown = "verifying", 70, true
	}
	switch {
	case strings.Contains(joined, "service stopped"):
		stage, progress, progressKnown = "stopping", 82, true
	case strings.Contains(joined, "docker compose -f"):
		stage, progress, progressKnown = "recreating", 82, true
	}
	switch {
	case strings.Contains(joined, "binary replaced"):
		stage, progress, progressKnown = "replacing", 92, true
	}
	switch {
	case strings.Contains(joined, "service started"):
		stage, progress, progressKnown = "restarting", 97, true
	}

	if status == "success" {
		return "completed", 100, true
	}
	if status == "failed" {
		if !progressKnown {
			progress = 0
		}
		return "failed", progress, progressKnown
	}
	if status == "running" && !progressKnown {
		if stage == "pending" {
			stage = "running"
		}
		return stage, 0, false
	}
	return stage, progress, progressKnown
}

func extractCurrentRunLines(lines []string) []string {
	lastStart := -1
	for i, line := range lines {
		if strings.Contains(line, "binary update started") || strings.Contains(line, "docker update started") {
			lastStart = i
		}
	}
	if lastStart >= 0 {
		return lines[lastStart:]
	}
	return lines
}

func extractProgress(lines []string) (int, bool) {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "PROGRESS:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "PROGRESS:"))
		n, err := strconv.Atoi(value)
		if err != nil {
			continue
		}
		if n < 0 {
			n = 0
		}
		if n > 100 {
			n = 100
		}
		return n, true
	}
	return 0, false
}
