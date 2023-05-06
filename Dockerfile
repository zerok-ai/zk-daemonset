FROM --platform=linux/amd64 golang:1.19.1-alpine3.16 AS build 
ENV GO111MODULE on
ENV CGO_ENABLED 0

RUN apk add make

WORKDIR /go/src/zerok-deamonset
ADD . .
RUN make build

FROM alpine:3.17
WORKDIR /zerok-deamonset
COPY --from=build /go/src/zerok-deamonset/zerok-deamonset .
COPY internal/config/config.yaml /zerok-deamonset/
CMD ["/zerok-deamonset/zerok-deamonset", "-c", "/zerok-deamonset/config.yaml"]