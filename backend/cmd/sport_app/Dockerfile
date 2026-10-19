ARG GO_VERSION=1.25.5
FROM golang:${GO_VERSION}-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/sport_app/exe /app/cmd/sport_app

FROM alpine:3.23
# Иначе time.LoadLocation("Europe/Moscow") падает → UTC и сутки дневника расходятся с consumed_at по полю day.
RUN apk add --no-cache tzdata ca-certificates wget && adduser -D -u 10001 sportik
ENV TZ=Europe/Moscow
WORKDIR /app
COPY --from=builder /app/cmd/sport_app/exe /app
COPY --from=builder /app/static /app/static
RUN mkdir /app/logs && chown sportik:sportik /app/logs
EXPOSE 5050
CMD ["/app/exe"]
