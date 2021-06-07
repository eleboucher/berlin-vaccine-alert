FROM golang:1.16
WORKDIR /root/
RUN go get -v github.com/rubenv/sql-migrate/...
COPY . .
RUN CGO_ENABLED=0 go build -o app .

RUN curl --location --silent --show-error --fail \
    https://github.com/Barzahlen/waitforservices/releases/download/v0.6/waitforservices \
    > /usr/local/bin/waitforservices && \
    chmod +x /usr/local/bin/waitforservices