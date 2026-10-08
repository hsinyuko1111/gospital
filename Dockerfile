FROM golang:1.25.1-alpine AS build
RUN go install github.com/swaggo/swag/cmd/swag@latest
WORKDIR /app
COPY . .
RUN go mod download
RUN swag init -g cmd/api/main.go -o docs
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-s -w" \
  -o /app/api \
  ./cmd/api

FROM alpine:latest AS deploy
WORKDIR /app
COPY --from=build /app/api .
COPY --from=build /app/docs ./docs
RUN addgroup -g 1000 appuser && \
  adduser -D -u 1000 -G appuser appuser
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
CMD ["./api"]