FROM golang:1.16-buster
MAINTAINER LiangMingYao <weixin:milkbobo>

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn
ENV BEEGO_RUNMODE prod

RUN apt-get update && apt-get install -y ca-certificates llvm clang mesa-opencl-icd ocl-icd-opencl-dev jq hwloc libhwloc-dev git vim

RUN mkdir /app
COPY . /app
WORKDIR /app/mes3
RUN git submodule init
RUN git submodule update --recursive
RUN go mod download -x
RUN make build-deps
RUN go build
ENTRYPOINT ["go","run","main.go"]
