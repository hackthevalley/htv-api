FROM golang:1.13 as builder
WORKDIR /app
COPY . /app
ENV GOPROXY "https://proxy.golang.org"
ENV GOSUMDB "sum.golang.org"
RUN env && go mod download && go mod verify && go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -mod vendor -ldflags="-w -s" -o app "github.com/hackthevalley/htv-api/server" && ls -la

FROM alpine:latest
RUN apk --no-cache add ca-certificates mailcap && addgroup -S app && adduser -S app -G app
USER app
WORKDIR /app
COPY --from=builder /app/app .
RUN ls -la /app/
ENTRYPOINT ["./app"]