# Build frontend
FROM node:16 AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm install
COPY web/ ./
RUN npm run build

# Build backend
FROM golang:1.19 AS backend-builder
WORKDIR /app
COPY . .
COPY --from=frontend-builder /app/web/dist ./internal/static
RUN go build -o bin/markdown-editor cmd/server/main.go

# Final stage
FROM debian:buster-slim
WORKDIR /app
COPY --from=backend-builder /app/bin/markdown-editor ./
COPY --from=backend-builder /app/internal/static ./static

ENV GO_ENV=production
EXPOSE 8080

CMD ["./markdown-editor"]
