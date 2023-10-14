FROM golang:1.18-alpine
WORKDIR /zk
COPY bin/zk-daemonset-amd64 /zk/zk-daemonset-amd64
COPY bin/zk-daemonset-arm64 /zk/zk-daemonset-arm64

CMD ["./app-start.sh","-amd64","zk-daemonset-amd64","-arm64","zk-daemonset-arm64","-c","/opt/config.yaml"]