FROM alpine:latest as builder

WORKDIR /app
RUN apk update && apk add go

COPY . .

RUN go build -o server ./cmd/web/*

FROM alpine:latest
RUN apk update && apk add typst fontconfig

COPY --from=builder /app/server ./
COPY --from=builder /app/static/JetBrainsMono-NFM.ttf /usr/share/fonts/truetype/

RUN fc-cache -f -v
RUN fc-list | grep "Jet"

EXPOSE 8080
ENTRYPOINT ["./server"]
