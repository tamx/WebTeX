FROM golang:1.21-bookworm as builder

WORKDIR /go/src/WebTex
COPY ./src/webtex/ .
RUN CGO_ENABLED=0 GOOS=linux go build -v -o webtex

FROM marketplace.gcr.io/google/debian12:latest

WORKDIR /usr/local/bin/
RUN apt-get update \
    && apt-get upgrade -y \
    && apt-get install -y fonts-takao texlive texlive-lang-cjk texlive-extra-utils \
    && mkdir tmp
COPY --from=builder /go/src/WebTex/webtex /usr/local/bin/
COPY ./src/webtex/static/ /usr/local/bin/static/

ENV TZ=Asia/Tokyo
ENTRYPOINT ["/usr/local/bin/webtex"]
