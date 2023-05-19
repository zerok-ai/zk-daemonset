FROM golang:1.18-alpine
WORKDIR /zk
COPY zk-daemonset .
COPY config.yaml .
CMD ["/zk/zk-daemonset", "-c", "/zk/config.yaml"]

