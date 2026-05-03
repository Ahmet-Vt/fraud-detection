FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl && adduser -D -u 65532 appuser
WORKDIR /app
RUN touch .env
COPY --from=build /out/api /app/api
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/api"]