FROM golang as base

WORKDIR /workdir/
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" ./cmd/notice

FROM busybox
LABEL maintainer="lwzm@qq.com"

COPY --from=base /workdir/notice /bin/notice

ENV PORT=80
EXPOSE 80

CMD [ "notice" ]
