# Simplified build for Favus CLI only
FROM golang:1.24.1-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o favus ./cmd

FROM python:3.10-slim AS runtime
WORKDIR /app

# Install Python dependencies
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

# Copy Go binary
COPY --from=go-builder /app/favus /usr/local/bin/favus

# Copy internal directory
COPY internal/ ./internal/

EXPOSE 8765

# Default command (can be overridden in docker-compose)
CMD ["python", "internal/wsserver/server.py"]
