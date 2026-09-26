# syntax=docker/dockerfile:1

FROM node:24-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.27-alpine AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/web/dist ./web/dist
ARG VERSION=dev
RUN go build -trimpath -ldflags "-s -w -X github.com/melojms/mealheiro/internal/api.Version=${VERSION}" -o /out/mealheiro ./cmd/mealheiro

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mealheiro /mealheiro
ENV DATA_DIR=/data ADDR=:8080 TZ=Europe/Lisbon
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/mealheiro", "healthcheck"]
ENTRYPOINT ["/mealheiro"]
