FROM alpine:latest as builder

WORKDIR /app
RUN apk update && apk add go

COPY . .

RUN go build -o server ./cmd/web/*

FROM alpine:latest
RUN apk update && apk add typst fontconfig
EXPOSE 8080
COPY --from=builder /app/server ./
COPY --from=builder /app/static/JetBrainsMono-NFM.ttf /usr/local/share/fonts/
RUN fc-cache -f -v
ENTRYPOINT ["./server"]
