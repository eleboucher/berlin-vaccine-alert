FROM golang:1.16-alpine
RUN apk --no-cache add ca-certificates sqlite gcc musl-dev
WORKDIR /root/
RUN go get -v github.com/rubenv/sql-migrate/...
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o app .
RUN sql-migrate up -env production
ENTRYPOINT ["./app", "run"]