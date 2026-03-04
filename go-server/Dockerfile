FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build


# DO NOT CHANGE THE VERSION NUMBER BEFORE CHECKING IF IT COMPILES WITH AN OLDER VERSION.
# LORIS NEEDS TO BE ON THE LATEST VERSION OF GO. BECAUSE I AM ON THE LATEST VERSION OF GO.

FROM golang:1.25-alpine AS backend-builder 
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=frontend-builder /app/frontend/dist ./static
EXPOSE 8080
ENTRYPOINT ["./server"]
