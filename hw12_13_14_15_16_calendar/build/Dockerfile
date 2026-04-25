# Собираем в гошке
FROM golang:1.25-alpine AS build

ENV BIN_FILE=/opt/calendar/calendar-app
ENV CODE_DIR=/go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY api/ ./api/
COPY migrations/ ./migrations/

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} ./cmd/calendar

# На выходе тонкий образ
FROM alpine:3.21

LABEL ORGANIZATION="OTUS Online Education"
LABEL SERVICE="calendar"
LABEL MAINTAINERS="student@otus.ru"

# Обновляем пакеты и устанавливаем tzdata для временных зон
RUN apk --no-cache update && \
    apk --no-cache upgrade && \
    apk --no-cache add tzdata ca-certificates && \
    rm -rf /var/cache/apk/*

# Создаем non-root пользователя
RUN addgroup -g 1000 calendar && \
    adduser -D -u 1000 -G calendar calendar

ENV BIN_FILE="/opt/calendar/calendar-app"
ENV CONFIG_FILE="/etc/calendar/calendar_config.yml"

# Создаем директории и устанавливаем права
RUN mkdir -p /opt/calendar /etc/calendar && \
    chown -R calendar:calendar /opt/calendar /etc/calendar

# Копируем бинарник и конфиг
COPY --from=build --chown=calendar:calendar ${BIN_FILE} ${BIN_FILE}
COPY --chown=calendar:calendar ./configs/calendar_config.yml ${CONFIG_FILE}

# Переключаемся на non-root пользователя
USER calendar

# HEALTHCHECK (проверяет, что приложение отвечает на HTTP/GRPC порт)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Используем exec форму для лучшей обработки сигналов
CMD ["/opt/calendar/calendar-app", "-config", "/etc/calendar/calendar_config.yml"]
