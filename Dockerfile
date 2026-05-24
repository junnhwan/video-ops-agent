# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.1
ARG NODE_VERSION=22
ARG GOPROXY=https://goproxy.cn,direct
ARG GOSUMDB=sum.golang.google.cn

# ---- build backend ----
FROM golang:${GO_VERSION}-alpine AS backend
ARG GOPROXY
ARG GOSUMDB
ENV GOPROXY=${GOPROXY} GOSUMDB=${GOSUMDB} CGO_ENABLED=0
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- build frontend ----
FROM node:${NODE_VERSION}-alpine AS frontend
WORKDIR /src
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ .
RUN npm run build

# ---- runtime ----
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata nginx && adduser -D -H -s /sbin/nologin app
WORKDIR /app

COPY --from=backend /out/server /app/server
COPY --from=frontend /src/dist /app/web/dist
COPY configs/ /app/configs/
COPY nginx.conf /etc/nginx/http.d/default.conf

RUN mkdir -p /app/data && chown -R app:app /app

EXPOSE 8080
CMD ["sh", "-c", "nginx && /app/server"]
