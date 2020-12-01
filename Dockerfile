FROM golang:1.15-alpine
RUN go get github.com/terraform-providers/terraform-provider-azurerm
RUN go get github.com/njucz/terraform-provider-azurerm-analysis
WORKDIR $GOPATH/src/github.com/njucz/terraform-provider-azurerm-analysis
RUN go install -o extract cmd/extract/main.go