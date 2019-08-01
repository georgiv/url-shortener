FROM golang:1.12.7

WORKDIR /url-shortener

COPY . .

RUN go get -d -v ./...
RUN go build main.go

EXPOSE 8888

CMD ["./main", "start"]
