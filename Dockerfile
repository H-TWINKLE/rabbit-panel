FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./

RUN npm ci

COPY frontend/ ./

RUN npm run build

FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

RUN apk add  gcc musl-dev

COPY backend/go.mod backend/go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct

RUN go mod download

COPY backend/ ./

COPY --from=frontend-builder /app/backend/dist ./dist

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o rabbit-panel .

FROM alpine:latest

WORKDIR /app

RUN apk add \
    docker-cli \
    docker-cli-compose \
    tzdata \
    ca-certificates

ENV TZ=Asia/Shanghai

COPY --from=backend-builder /app/rabbit-panel .

RUN mkdir -p /app/compose_projects

EXPOSE 9999

CMD ["./rabbit-panel"]
