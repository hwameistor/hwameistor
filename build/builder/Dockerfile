FROM golang:alpine

# set mirror repository for the package management
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/' /etc/apk/repositories
RUN apk add make docker-cli git build-base bash

COPY ./tools/dev/operator-sdk-v0.18.2-x86_64-linux-gnu /usr/bin/operator-sdk
RUN chmod +x /usr/bin/operator-sdk

ENV GO111MODULE=on GOPROXY=https://goproxy.io

RUN git clone -b v0.18.6 --single-branch https://github.com/kubernetes/code-generator && mv code-generator /code-generator
