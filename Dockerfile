FROM golang:1.22.3-alpine3.20 AS builder
RUN apk add git gcc musl-dev linux-headers
WORKDIR /opt/build
COPY .git /tmp/build/.git
RUN git clone /tmp/build /opt/build && rm -rf /tmp/build
RUN go get
RUN go build
RUN go build ./cli/satdress-cli.go

FROM alpine:3.14
VOLUME /var/lib/satdress
RUN apk add curl jq
COPY --from=builder /opt/build/satdress /opt/build/satdress-cli /usr/local/bin/
EXPOSE 8080
ENTRYPOINT /usr/local/bin/satdress --conf=/run/secrets/satdress.yml