FROM golang:1.16-buster as builder

WORKDIR /go/src/WebTex
COPY ./src/webtex/ .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o webtex

FROM marketplace.gcr.io/google/debian10:latest

WORKDIR /usr/local/bin/
RUN apt-get update \
    && apt-get install -y fonts-takao texlive texlive-lang-cjk texlive-extra-utils
COPY --from=builder /go/src/WebTex/webtex /usr/local/bin/
COPY ./src/webtex/index.html /usr/local/bin/

ENV TZ=Asia/Tokyo
ENTRYPOINT ["/usr/local/bin/webtex"]
