# Stage 1: Build the Go application
FROM golang:1.23 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /build

# Копируем файлы зависимостей для скачивания модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем оставшиеся файлы проекта
COPY . .

# Собираем бинарный файл в корневой директории /build
RUN go build -o main ./cmd/main.go

# Stage 2: Create a lightweight image with the application binary
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из предыдущего этапа
COPY --from=builder /build/main /app/main

# Устанавливаем точку входа для контейнера
CMD ["/app/main"]