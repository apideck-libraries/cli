FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X main.version=$(cat VERSION 2>/dev/null || echo dev)" -o /apideck ./cmd/apideck

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /apideck /apideck
ENTRYPOINT ["/apideck"]
