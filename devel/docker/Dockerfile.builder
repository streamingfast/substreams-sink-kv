FROM --platform=linux/amd64 ubuntu:20.04 AS x86_64-linux-gnu

FROM goreleaser/goreleaser-cross:v1.20.5

RUN cd /root &&\
    mkdir wasmedge-ubuntu-20.04-x86-64 &&\
    cd wasmedge-ubuntu-20.04-x86-64 &&\
    wget -O wasmedge.tar.gz https://github.com/WasmEdge/WasmEdge/releases/download/0.11.2/WasmEdge-0.11.2-ubuntu20.04_x86_64.tar.gz &&\
    tar -xzvf wasmedge.tar.gz &&\
    mkdir -p /usr/x86_64-linux-gnu/lib &&\
    cp -R WasmEdge-0.11.2-Linux/include/* /usr/x86_64-linux-gnu/include &&\
    cp -R WasmEdge-0.11.2-Linux/lib/* /usr/x86_64-linux-gnu/lib &&\
    rm -rf wasmedge-ubuntu-20.04-x86-64

RUN cd /root &&\
    mkdir wasmedge-ubuntu-20.04-aarch64 &&\
    cd wasmedge-ubuntu-20.04-aarch64 &&\
    wget -O wasmedge.tar.gz https://github.com/WasmEdge/WasmEdge/releases/download/0.11.2/WasmEdge-0.11.2-manylinux2014_aarch64.tar.gz &&\
    tar -xzvf wasmedge.tar.gz &&\
    mkdir -p /usr/lib &&\
    cp -R WasmEdge-0.11.2-Linux/include/* /usr/include &&\
    cp -R WasmEdge-0.11.2-Linux/lib64/* /usr/lib &&\
    rm -rf wasmedge-ubuntu-20.04-aarch64

RUN cd /root &&\
    mkdir wasmedge-darwin-arm64 &&\
    cd wasmedge-darwin-arm64 &&\
    wget -O wasmedge.tar.gz https://github.com/WasmEdge/WasmEdge/releases/download/0.11.2/WasmEdge-0.11.2-darwin_arm64.tar.gz &&\
    tar -xzvf wasmedge.tar.gz &&\
    mkdir -p /usr/local/osxcross/{lib,include}/arm64 &&\
    cp -Ra WasmEdge-0.11.2-Darwin/include/* /usr/local/osxcross/include/arm64 &&\
    cp -Ra WasmEdge-0.11.2-Darwin/lib/libwasmedge.* /usr/local/osxcross/lib/arm64 &&\
    rm -rf wasmedge-darwin-arm64

RUN cd /root &&\
    mkdir wasmedge-darwin-amd64 &&\
    cd wasmedge-darwin-amd64 &&\
    wget -O wasmedge.tar.gz https://github.com/WasmEdge/WasmEdge/releases/download/0.11.2/WasmEdge-0.11.2-darwin_x86_64.tar.gz &&\
    tar -xzvf wasmedge.tar.gz &&\
    mkdir -p /usr/local/osxcross/{lib,include}/amd64 &&\
    cp -R WasmEdge-0.11.2-Darwin/include/* /usr/local/osxcross/include/amd64 &&\
    cp -Ra WasmEdge-0.11.2-Darwin/lib/libwasmedge.* /usr/local/osxcross/lib/amd64 &&\
    rm -rf wasmedge-darwin-amd64

COPY --from=x86_64-linux-gnu /lib/x86_64-linux-gnu/libz.so.1* /usr/x86_64-linux-gnu/lib
COPY --from=x86_64-linux-gnu /lib/x86_64-linux-gnu/libtinfo.so.6* /usr/x86_64-linux-gnu/lib

