FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY . .


RUN go mod init host_story_api || true
RUN go get github.com/docker/docker@v28.5.2+incompatible \
    github.com/docker/go-connections@v0.6.0 \
    github.com/golang-jwt/jwt/v4@v4.5.2

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api .

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /out/api /app/api

EXPOSE 8082
CMD ["/app/api"]