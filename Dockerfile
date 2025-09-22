FROM alpine:latest as builder

WORKDIR /app
RUN apk update && apk add go tzdata

COPY . .

RUN go build -o server ./cmd/web/*

FROM alpine:latest
RUN apk update && apk add typst fontconfig

COPY --from=builder /app/server ./
COPY --from=builder /app/static/fonts/* /usr/share/fonts/truetype/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

RUN fc-cache -f -v
ENV TZ=America/Sao_Paulo

EXPOSE 8080
ENTRYPOINT ["./server"]
