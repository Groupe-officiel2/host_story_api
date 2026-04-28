FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY *.go .

RUN go mod init host_story_api
RUN go get github.com/docker/docker@v28.5.2+incompatible \
    github.com/docker/go-connections@v0.6.0 \
    github.com/golang-jwt/jwt/v4@v4.5.2 \
    github.com/containerd/errdefs@latest \
    github.com/moby/docker-image-spec/specs-go/v1@latest \
    go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp@latest \
    go.opentelemetry.io/otel@latest \ 
    github.com/joho/godotenv@v1.5.1 \ 
    github.com/ovh/go-ovh@v1.9.0
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/api .

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /out/api /app/api

EXPOSE 8082
CMD ["/app/api"]

