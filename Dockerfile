FROM golang:1.16.6-alpine AS builder

WORKDIR /repo

ADD go.mod .
ADD go.sum .

ADD . ./

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache bash tzdata

ENV TZ="Asia/Shanghai"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -mod vendor -o api cmd/main.go

FROM chromedp/headless-shell:latest
ENV TZ="Asia/Shanghai"
RUN apt-get update -y && apt install dumb-init
ENTRYPOINT ["dumb-init", "--"]
WORKDIR /repo
COPY --from=builder /repo/api api
COPY --from=builder /repo/.env .env
CMD ["/repo/api"]