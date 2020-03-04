FROM golang as base

LABEL maintainer="lwzm@qq.com"

WORKDIR /workdir
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o o ./agency

FROM busybox
COPY --from=base /workdir/o /agency

ENV PORT=80
EXPOSE 80

ENTRYPOINT [ "/agency" ]
