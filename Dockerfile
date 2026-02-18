# Stage 1: Build Vue frontend
FROM node:20-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build Go backend
FROM golang:1.22-alpine AS backend-build
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum* ./
RUN go mod download 2>/dev/null || true
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# Stage 3: Production runtime
FROM alpine:3.19

# Install ffmpeg
RUN apk add --no-cache ffmpeg ca-certificates

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy built artifacts
COPY --from=backend-build /app/backend/server /app/server
COPY --from=frontend-build /app/frontend/dist /app/static

# Create conversions directory
RUN mkdir -p /app/conversions && chown appuser:appgroup /app/conversions

WORKDIR /app

# Environment configuration
ENV PORT=8080
ENV PROXY_HEADER=X-Forwarded-For
ENV MAX_UPLOAD_MB=200
ENV STATIC_DIR=/app/static
ENV CONVERSIONS_DIR=/app/conversions

EXPOSE 8080

USER appuser

CMD ["./server"]
