FROM golang:1.26.0-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/sport_app/exe /app/cmd/sport_app

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /app/cmd/sport_app/exe /app
CMD ["/app/exe"]