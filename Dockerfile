# Build stage for frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Build stage for backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
COPY backend/ .
COPY --from=frontend-builder /app/dist/assets ./static/assets
COPY --from=frontend-builder /app/dist/manifest.json ./static/manifest.json
RUN go build -o server main.go

# Final stage
FROM alpine:latest
WORKDIR /app
COPY --from=backend-builder /app/server .
COPY --from=backend-builder /app/static ./static
COPY --from=backend-builder /app/index.html .

# Default markdown directory inside container
ENV MARKDOWN_DIR=/data/markdown
# Create the default markdown directory
RUN mkdir -p ${MARKDOWN_DIR}

EXPOSE 8080

CMD ["./server"]
