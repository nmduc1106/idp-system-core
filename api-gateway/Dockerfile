# Bước 1: Build ứng dụng
FROM golang:1.25-alpine AS builder
# Cài đặt các công cụ cần thiết
RUN apk add --no-cache git

# Thiết lập thư mục làm việc
WORKDIR /app

# Copy các file quản lý thư viện
COPY go.mod go.sum ./
RUN go mod download

# Copy toàn bộ code
COPY . .

# Build ứng dụng thành file thực thi 'main'
RUN go build -o main ./cmd/api

# Bước 2: Chạy ứng dụng (Dùng image nhẹ để giảm dung lượng)
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy file thực thi từ bước builder
COPY --from=builder /app/main .
# Copy thư mục docs (nếu bạn dùng Swagger)
COPY --from=builder /app/docs ./docs

# Mở cổng 8080
EXPOSE 8080

# Lệnh chạy ứng dụng
CMD ["./main"]