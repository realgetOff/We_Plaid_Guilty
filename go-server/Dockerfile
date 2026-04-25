ARG GO_ARCH=arm64
ARG FINAL_PLATFORM=linux/arm64

# --- Étape de build Frontend ---
FROM --platform=linux/amd64 node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# --- Étape de build Backend ---
FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder
# On récupère l'argument d'architecture déclaré au tout début
ARG GO_ARCH
RUN apk add --no-cache ca-certificates
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GO_ARCH} go build -o /app/server .

# --- Étape SonarQube (Ne polluera pas l'image finale) ---
FROM sonarsource/sonar-scanner-cli:latest AS sonarqube
WORKDIR /usr/src
COPY frontend/ /usr/src/frontend/
COPY backend/ /usr/src/backend/
# (Optionnel) Tu peux rajouter tes configs Sonar ici
COPY sonar-project.properties /usr/src/

# --- Image Finale ---
ARG FINAL_PLATFORM
FROM --platform=${FINAL_PLATFORM} alpine:3.19
WORKDIR /app
COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./static
EXPOSE 8080
ENTRYPOINT ["./server"]
