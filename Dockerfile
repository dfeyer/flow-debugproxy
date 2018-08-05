FROM golang:latest

WORKDIR /go/src/flow-debugproxy
COPY . .

RUN go get -v
RUN go install -v

CMD ["flow-debugproxy"]
