FROM golang:1.13 AS builder
WORKDIR /app
COPY . /app
ENV GOPROXY "https://proxy.golang.org"
ENV GOSUMDB "sum.golang.org"
ENV GO111MODULE=on
RUN env && go mod download && go mod tidy && go mod verify && go mod vendor
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -mod vendor -ldflags="-w -s" -o app "github.com/hackthevalley/htv-api/server"

FROM alpine:latest AS final
RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /app/app .
RUN ls -la /app/
ENTRYPOINT ["./app"]