# syntax=docker/dockerfile:1
FROM golang:1.16
WORKDIR /go/src/github.com/bluestoneag/post-requester/
RUN go get -d -v golang.org/x/net/html  
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o post-requester .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/bluestoneag/post-requester/post-requester ./
CMD ["./post-requester"]  