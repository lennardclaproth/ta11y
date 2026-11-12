# Production image (Cloud Run): the Go API, serving the built web app at / and itself under /api.
#   docker build -t ta11y .

FROM node:22-slim AS web
WORKDIR /web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
# Same origin as the API, which serves this build: a relative base keeps the image domain-agnostic.
ENV VITE_USE_MOCKS=false VITE_API_URL=/api
RUN npm run build

FROM golang:1.25 AS api
WORKDIR /src
# The repo-root go.work is not in the image; build the module on its own.
ENV GOWORK=off CGO_ENABLED=0 GOOS=linux
COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download
COPY apps/api/ ./
RUN go build -trimpath -ldflags="-s -w" -o /out/ta11y ./cmd/ta11y

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /out/ta11y /app/ta11y
COPY --from=web /web/build /app/web
COPY apps/api/config.cloudrun.yaml /app/config.cloudrun.yaml

ENV CONFIG_PATH=/app/config.cloudrun.yaml
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/ta11y"]
