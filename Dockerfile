FROM golang:latest

WORKDIR /go/src/github.com/dfeyer/flow-debugproxy
COPY . .

RUN go get -v
RUN go install -v

ENTRYPOINT ["flow-debugproxy"]
