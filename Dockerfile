# Stage 1: build
FROM w32blaster/go-govendor as builder

RUN mkdir -p /go/src/github.com/w32blaster/hsl-gfts-parser/vendor

ENV GOPATH=/go/
WORKDIR /go/src/github.com/w32blaster/hsl-gfts-parser/


ADD ./*.go /go/src/github.com/w32blaster/hsl-gfts-parser/
ADD ./vendor/vendor.json /go/src/github.com/w32blaster/hsl-gfts-parser/vendor/vendor.json

RUN apk update && apk add --no-cache gcc g++ musl-dev
RUN cd /go/src/github.com/w32blaster/hsl-gfts-parser/ && \
    govendor list && \
    govendor fetch -v +out

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o hsl-parser .

# Stage 2: runtime container
FROM alpine:latest

COPY --from=builder /go/src/github.com/w32blaster/hsl-gfts-parser/hsl-parser /root/

RUN apk update && \

    # install SQlite3 to set up a new database
    apk add --no-cache sqlite wget ca-certificates && \

    # clean up
    rm -rf /tmp/* && \
    rm -rf /var/cache/apk/* && \

    mkdir -p /root/db && \

    # make parser runnable
    chmod +x /root/hsl-parser

WORKDIR /root
CMD ["./hsl-parser"]
