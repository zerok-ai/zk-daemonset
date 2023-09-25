FROM golang:1.18-alpine
WORKDIR /zk
COPY zk-daemonset /zk/zk-daemonset
CMD ["dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/zk/zk-daemonset", "-c", "/zk/config/config.yaml"]

