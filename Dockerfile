# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG TARGETOS TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/server ./cmd/server

FROM node:22-alpine AS frontend
WORKDIR /web
COPY web/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY web .
RUN npm run desk:bundle

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app -u 10001 app && mkdir -p /app/var/files && chown -R app:app /app
WORKDIR /app
COPY --from=backend /out/server /app/server
COPY --from=frontend /web/dist /app/web/dist
COPY migrations /app/migrations
USER app
EXPOSE 8080
ENTRYPOINT ["/app/server"]
