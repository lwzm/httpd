FROM golang as base

LABEL maintainer="lwzm@qq.com"

ARG sub=stream

WORKDIR /workdir
COPY ./ ./
RUN echo $sub && CGO_ENABLED=0 go build -ldflags "-s -w" -o httpd ./$sub

FROM scratch
COPY --from=base /workdir/httpd /

ENV PORT=80
EXPOSE 80

ENTRYPOINT [ "/httpd" ]
