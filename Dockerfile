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
RUN go build -trimpath -ldflags "-s -w -X github.com/melojms/mm-budget/internal/api.Version=${VERSION}" -o /out/mm-budget ./cmd/mm-budget

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mm-budget /mm-budget
ENV DATA_DIR=/data ADDR=:8080 TZ=Europe/Lisbon
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD ["/mm-budget", "healthcheck"]
ENTRYPOINT ["/mm-budget"]
