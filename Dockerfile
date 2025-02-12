# Stage 1: Сборка приложения
FROM golang:1.22 as builder

WORKDIR /app

# Копируем только файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости с использованием кэша
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Копируем весь код
COPY . .

# Собираем приложение
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /avito-shop ./cmd/app/main.go

# Stage 2: Финальный контейнер
FROM scratch

WORKDIR /root/

# Копируем бинарник из builder stage
COPY --from=builder /avito-shop .

# Прокидываем переменные окружения через .env (если потребуется)
ENV SERVER_PORT=8080

# Открываем порт
EXPOSE 8080

# Запускаем сервис
CMD ["./avito-shop"]