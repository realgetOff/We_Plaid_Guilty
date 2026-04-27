ARG GO_ARCH=arm64
ARG FINAL_PLATFORM=linux/arm64

# checkov:skip=CKV_DOCKER_3:builder stage, not final image
FROM --platform=linux/amd64 node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# checkov:skip=CKV_DOCKER_3:builder stage, not final image
FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder
ARG GO_ARCH
RUN apk add --no-cache ca-certificates
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GO_ARCH} go build -o /app/server .

# checkov:skip=CKV_DOCKER_7:sonarqube stage uses latest by design
FROM sonarsource/sonar-scanner-cli:latest AS sonarqube
WORKDIR /usr/src
COPY frontend/ /usr/src/frontend/
COPY backend/ /usr/src/backend/
COPY sonar-project.properties /usr/src/

ARG FINAL_PLATFORM
FROM --platform=${FINAL_PLATFORM} alpine:3.19 AS final
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./static
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
  CMD wget -qO- http://localhost:8081/health || exit 1
ENTRYPOINT ["./server"]
ENTRYPOINT ["./server"]
