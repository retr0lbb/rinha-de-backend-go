FROM golang:1.24-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o server ./src

# imagem final mínima
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server .

COPY files ./files

EXPOSE 9999

CMD [ "./server" ]