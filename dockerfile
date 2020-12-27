FROM golang:alpine as build-env

ENV GO111MODULE=on

RUN apk update && apk add bash ca-certificates git gcc g++ libc-dev

RUN mkdir /docker_chat_grpc
RUN mkdir -p /docker_chat_grpc/protobuff

WORKDIR /docker_chat_grpc

COPY ./protobuff/service.pb.go /docker_chat_grpc/protobuff
COPY ./main.go /docker_chat_grpc

COPY go.mod .
COPY go.sum .

RUN go mod download

RUN go build -o docker_chat_grpc .

CMD ./docker_chat_grpc