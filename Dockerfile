FROM golang

ADD . /go/src/github.com/rahulroshan96/proxy2/

WORKDIR /go/src/github.com/rahulroshan96/proxy2/



WORKDIR /go/src/github.com/rahulroshan96/proxy2/cmd

RUN go get ./

RUN go build main.go

EXPOSE 4996

EXPOSE 5996

CMD ["./main"]