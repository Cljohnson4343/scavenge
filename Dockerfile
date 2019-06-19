FROM golang

RUN go get -v github.com/cljohnson4343/scavenge
RUN go install -i github.com/cljohnson4343/scavenge

ADD config.yaml /go/src/github.com/cljohnson4343/scavenge

CMD ["/go/bin/scavenge", "serve"]

EXPOSE 4343

