# [WIP] smaller docker image
###########################
# STEP 1 build executable binary
############################
# golang alpine 1.12.6
FROM golang@sha256:eec8b6c0bc53eff8fc6d5f934279138854f6c93c7d997cb292bcab09d3c6a3b6 as builder
ENV GOPROXY "https://proxy.golang.org"
ENV GOSUMDB "sum.golang.org"
ENV GO111MODULE=on
# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /app
COPY . /app

# Fetch dependencies.
RUN env && go mod download && go mod tidy && go mod verify && go mod vendor

# Build the binary
#RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o /go/bin/hello .
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -mod vendor -ldflags="-w -s" -a -o app "github.com/hackthevalley/htv-api/server"

############################
# STEP 2 build a small image
############################
FROM scratch

# Import from builder.
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy our static executable
COPY --from=builder /app/app .

# Use an unprivileged user.
USER appuser

# Run the binary.
ENTRYPOINT ["./app"]
