FROM golang:1.19
ENV GO111MODULE=on

RUN ln -snf /usr/share/zoneinfo/Asia/Almaty /etc/localtime && echo Asia/Almaty > /etc/timezone
WORKDIR /app
COPY . .
RUN go build -o app ./cmd/server
CMD ["./app"]
