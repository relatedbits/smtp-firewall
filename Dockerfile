FROM golang:1.21 as build

ENV GOCACHE "/tmp/.gocache"
WORKDIR /go/src/app
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target=/tmp/.gocache/ \
    CGO_ENABLED=0 go build -o /go/bin/smtpfirewall ./cmd/smtpfirewall

RUN curl --fail -o bad_domains.txt https://raw.githubusercontent.com/disposable-email-domains/disposable-email-domains/master/disposable_email_blocklist.conf

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY ./cmd/smtpfirewall/example.yaml default.yaml
COPY --from=build /go/src/app/bad_domains.txt .
COPY --from=build /go/bin/smtpfirewall .

CMD ["/app/smtpfirewall"]
