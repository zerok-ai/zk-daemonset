FROM golang:1.18 as builder
WORKDIR /workspace
COPY . .
RUN CGO_ENABLED=0 go build -a -o ../app main.go
ENTRYPOINT ["/app"]