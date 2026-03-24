# Build FansEdu API (Go)
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/fansedu-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/fansedu-api /usr/local/bin/fansedu-api
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/fansedu-api"]
