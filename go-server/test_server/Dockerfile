FROM --platform=linux/arm64 alpine:3.19
WORKDIR /app
COPY server_arm64 /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]
