FROM golang:1.20.3-alpine AS builder

WORKDIR /usr/src/app
COPY . . 

RUN go get github.com/gorilla/websocket

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-s' -o main .

FROM scratch

COPY --from=builder /usr/src/app/main /main
EXPOSE 3000
CMD [ "/main" ]
