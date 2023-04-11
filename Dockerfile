# WITH Go Modules
FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git

RUN mkdir $GOPATH/src/leave_reciever

ADD ./client.go $GOPATH/src/leave_reciever

WORKDIR $GOPATH/src/leave_reciever
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod init
RUN go mod tidy
RUN go mod download
RUN mkdir /pro
RUN go build -o /pro/client client.go

FROM alpine:latest
RUN mkdir /pro
COPY --from=builder /pro/client /pro/client
WORKDIR /pro
CMD ["/pro/client"]
