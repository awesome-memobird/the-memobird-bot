FROM golang:1.13

WORKDIR /bot
COPY . .

RUN go build -v -mod=vendor /bot/cmd/the-memobird-bot

CMD ["./the-memobird-bot"]
