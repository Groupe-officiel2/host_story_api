FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api ./cmd/api

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /out/api /app/api
COPY --from=builder /app/configs /app/configs

EXPOSE 8082
CMD ["/app/api"]

