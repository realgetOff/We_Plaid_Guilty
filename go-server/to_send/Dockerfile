FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=frontend /app/dist ./frontend/dist
EXPOSE 8080
CMD ["./server"]
