# stage 1

FROM golang:1.20-alpine3.17 as builder

RUN mkdir /build

ADD . /build

WORKDIR /build

RUN go build -o main .

CMD [ "ls" ]

CMD [ "pwd" ]

# stage 2

FROM alpine:3.17

RUN adduser -D newuser

USER newuser

COPY . /app

COPY --from=builder /build/main /build/.env /home/newuser

WORKDIR /home/newuser

EXPOSE 4000

CMD [ "ls" ]

CMD [ "pwd" ]

CMD [ "./main" ]
