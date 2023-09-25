FROM golang:1.18-alpine
WORKDIR /zk
COPY zk-daemonset /zk/zk-daemonset
#CMD ["/zk/zk-daemonset", "-c", "/zk/config/config.yaml"]

RUN "go install github.com/go-delve/delve/cmd/dlv@master"
CMD ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/zk/zk-daemonset", "-c", "/zk/config/config.yaml"]

