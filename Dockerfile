FROM node:20-alpine AS frontend-builder

RUN corepack enable && corepack prepare pnpm@latest --activate

WORKDIR /app/frontend

COPY frontend/package.json frontend/pnpm-lock.yaml ./

RUN pnpm config set registry https://registry.npmmirror.com && \
    pnpm install --frozen-lockfile

COPY frontend/ ./

RUN pnpm run build

FROM golang:1.24-alpine AS backend-builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

WORKDIR /app

COPY backend/go.mod backend/go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && \
    go mod download

COPY backend/ ./

COPY --from=frontend-builder /app/backend/.dist ./.dist

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildTime=${BUILD_TIME}'" -o rabbit-panel .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache \
    docker-cli \
    docker-cli-compose \
    tzdata \
    ca-certificates && \
    mkdir -p /app/compose_projects

ENV TZ=Asia/Shanghai

COPY --from=backend-builder /app/rabbit-panel .

EXPOSE 3958

CMD ["./rabbit-panel"]
