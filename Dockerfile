FROM golang:1.15.5-alpine AS builder
RUN apk update && apk add --no-cache bash git openssh
RUN set -ex && \
    apk add --no-cache gcc musl-dev

RUN set -ex && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/cc1obj && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/lto1 && \
    rm -f /usr/libexec/gcc/x86_64-alpine-linux-musl/6.4.0/lto-wrapper && \
    rm -f /usr/bin/x86_64-alpine-linux-musl-gcj

FROM builder AS builder1
RUN mkdir -p $GOPATH/src/github.com/terraform-providers && cd $GOPATH/src/github.com/terraform-providers && git clone --verbose -b master --single-branch https://github.com/terraform-providers/terraform-provider-azurerm.git
RUN mkdir -p $GOPATH/src/github.com/njucz && cd $GOPATH/src/github.com/njucz && git clone --verbose -b master --single-branch https://github.com/njuCZ/terraform-provider-azurerm-analysis.git
RUN cd $GOPATH/src/github.com/njucz/terraform-provider-azurerm-analysis && git pull && \
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o extract cmd/extract/main.go && \
        GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o analysis cmd/server/main.go && \
        mv extract analysis $GOPATH/bin/

FROM builder
ENV GOFLAGS -mod=vendor
ENV PROVIDER_REPO_PATH $GOPATH/src/github.com/terraform-providers/terraform-provider-azurerm
ENV EXTRACT_CMD_PATH $GOPATH/bin/extract
COPY --from=builder1 $GOPATH/src $GOPATH/src
COPY --from=builder1 $GOPATH/bin $GOPATH/bin
EXPOSE 8080
ENTRYPOINT $GOPATH/bin/analysis