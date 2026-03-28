# FROM --platform=linux/amd64 node:20-alpine AS frontend-builder
# WORKDIR /app/frontend
# COPY frontend/package*.json ./
# RUN npm install
# COPY frontend/ .
# RUN npm run build
# 
# FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder
# RUN apk add --no-cache ca-certificates
# WORKDIR /app/backend
# COPY backend/go.mod backend/go.sum ./
# RUN go mod download
# COPY backend/ .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server .
# 
# FROM --platform=linux/arm64 alpine:3.19
# WORKDIR /app
# COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=backend-builder /app/server .
# COPY --from=frontend-builder /app/frontend/dist ./static
# EXPOSE 8080
# ENTRYPOINT ["./server"]

# THE BELOW CODE IS THE PRODUCTION ENVIRONMENT, DONT TOUCH IT PLEASE

FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM golang:1.25-alpine AS backend-builder 
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server .

FROM --platform=linux/amd64 alpine:3.19
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./static
EXPOSE 8080
ENTRYPOINT ["./server"]

# FROM --platform=linux/amd64 node:20-alpine AS frontend-builder
# WORKDIR /app/frontend
# COPY frontend/package*.json ./
# RUN npm install
# COPY frontend/ .
# RUN npm run build
# 
# FROM --platform=linux/amd64 golang:1.25-alpine AS backend-builder
# RUN apk add --no-cache ca-certificates
# WORKDIR /app/backend
# COPY backend/go.mod backend/go.sum ./
# RUN go mod download
# COPY backend/ .
# RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /app/server .
# 
# FROM --platform=linux/arm64 alpine:3.19
# WORKDIR /app
# COPY --from=backend-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# COPY --from=backend-builder /app/server .
# COPY --from=frontend-builder /app/frontend/dist ./static
# EXPOSE 8080
# ENTRYPOINT ["./server"]
# 
# THE BELOW CODE IS THE PRODUCTION ENVIRONMENT, DONT TOUCH IT PLEASE

# FROM node:20-alpine AS frontend-builder
# WORKDIR /app/frontend
# COPY frontend/package*.json ./
# RUN npm install
# COPY frontend/ .
# RUN npm run build

# FROM golang:1.25-alpine AS backend-builder 
# WORKDIR /app/backend
# COPY backend/go.mod backend/go.sum ./
# RUN go mod download
# COPY backend/ .
# RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

# FROM alpine:3.19
# RUN apk add --no-cache ca-certificates
# WORKDIR /app
# COPY --from=backend-builder /app/server .
# COPY --from=frontend-builder /app/frontend/dist ./static
# EXPOSE 8080
# ENTRYPOINT ["./server"]
