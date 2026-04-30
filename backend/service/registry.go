package service

import (
	"net/http"
	"strings"

	"rabbit-panel/repository"
)

// RegistryService 镜像仓库服务
type RegistryService struct {
	fileRepo repository.IFileRepository
}

// NewRegistryService 创建仓库服务
func NewRegistryService(fr repository.IFileRepository) *RegistryService {
	return &RegistryService{fileRepo: fr}
}

// ListRegistries 列出仓库
func (s *RegistryService) ListRegistries() ([]*repository.RegistryRecord, error) {
	registries, err := s.fileRepo.LoadRegistries()
	if err != nil {
		return nil, err
	}

	result := make([]*repository.RegistryRecord, 0, len(registries))
	for _, r := range registries {
		// 清除密码
		r.Password = ""
		result = append(result, r)
	}
	return result, nil
}

// CreateRegistry 创建仓库
func (s *RegistryService) CreateRegistry(r *repository.RegistryRecord) error {
	registries, err := s.fileRepo.LoadRegistries()
	if err != nil {
		return err
	}
	registries[r.URL] = r
	return s.fileRepo.SaveRegistries(registries)
}

// DeleteRegistry 删除仓库
func (s *RegistryService) DeleteRegistry(url string) error {
	registries, err := s.fileRepo.LoadRegistries()
	if err != nil {
		return err
	}
	delete(registries, url)
	return s.fileRepo.SaveRegistries(registries)
}

// TestRegistry 测试仓库连接
func (s *RegistryService) TestRegistry(url, username, password string) (bool, string) {
	// 尝试 HTTPS，然后 HTTP
	for _, scheme := range []string{"https", "http"} {
		targetURL := scheme + "://" + strings.TrimPrefix(url, "http://") + "/v2/"

		req, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			continue
		}
		if username != "" {
			req.SetBasicAuth(username, password)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			return true, "连接成功"
		}
	}

	return false, "连接失败，请检查地址和凭据"
}