FROM golang:1.18-alpine
WORKDIR /zk
COPY zk-daemonset /zk/zk-daemonset
CMD ["/zk/zk-daemonset", "-c", "/zk/config/config.yaml"]

