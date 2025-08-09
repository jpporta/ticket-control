FROM golang:1.24 as builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY . .


RUN go mod tidy
RUN go build -o server ./cmd/web/*

FROM scratch
RUN apk update && apk add typst
EXPOSE 8080
COPY --from=builder /app/server /server
ENTRYPOINT ["./server"]
