FROM golang:1.16-buster
MAINTAINER LiangMingYao <weixin:milkbobo>

ENV GO111MODULE on
ENV GOPROXY https://goproxy.cn
ENV BEEGO_RUNMODE prod

ENV FULLNODE_API_INFO eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.cNKFaIl_ZOAoZf0wgFLvQ22BghXTEXPXydeer5WSths:ws://10.20.5.104:1234/rpc/v0
ENV FULLNODE_API_INFO2 eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.dIq2urUe6ukxLq629KdX7vj_Gu0Fxw4h9lCANUA1BME:ws://10.20.5.105:1234/rpc/v0

RUN apt-get update && apt-get install -y ca-certificates llvm clang mesa-opencl-icd ocl-icd-opencl-dev jq hwloc libhwloc-dev git vim

RUN mkdir /app
COPY . /app
WORKDIR /app/sell999
RUN git submodule init
RUN git submodule update --recursive
RUN go mod download -x
RUN make build-deps
RUN go build
ENTRYPOINT ["go","run","main.go"]
