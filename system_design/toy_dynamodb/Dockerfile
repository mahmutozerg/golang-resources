FROM golang:1.25.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN CGO_ENABLED=0 go build -o server ./cmd/server/main.go

# 2. Run Aşaması (Çalıştırma)
# Final imajımız temiz bir Alpine olsun (veya scratch)
FROM alpine:latest

WORKDIR /root/

# Derlenen binary'i buraya taşı
COPY --from=builder /app/server .

# gRPC portunu dışarıya duyur
EXPOSE 50051

# Uygulamayı başlat
CMD ["./server"]
