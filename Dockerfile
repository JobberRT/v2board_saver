FROM golang:1.19 AS build
WORKDIR /build
COPY . .
#Uncomment this if you're building this image in china mainland
#ENV GOPROXY "https://goproxy.cn"
#Uncomment this and change it to your own timezone to currect logging timestamp
ENV TZ "Asia/Shanghai"
RUN go mod tidy && \
    go mod vendor && \
    go build -o saver


FROM ubuntu:22.04 AS run
WORKDIR /
RUN apt-get update && \
    apt-get install --no-install-recommends -y tzdata ca-certificates iputils-ping curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
COPY ./config.example.yaml ./config.yaml
COPY --from=build /build/saver .
CMD ["./saver"]
