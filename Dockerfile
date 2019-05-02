FROM golang:1.12-alpine as build

RUN mkdir /app
WORKDIR /app
RUN apk add --update --no-cache git
COPY go.mod .
# https://github.com/golang/go/issues/27925 https://github.com/golang/go/issues/29278
#COPY go.sum .
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM scratch

WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/bin/app httpd
COPY image.png image.png

ENTRYPOINT ["/app/httpd", "--port", "80", "--file", "image.png", "--url", "/testimage"]