FROM golang:1.16
WORKDIR /go/src/github.com/eleboucher/berlin-vaccine-alert
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite gcc musl-dev
WORKDIR /root/
COPY --from=0 /go/src/github.com/eleboucher/berlin-vaccine-alert/app .
COPY  ./.config.yml .
CMD ["./app", "run"]