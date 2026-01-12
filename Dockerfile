# 使用 Ubuntu 18.04 基础镜像（GLIBC 2.27，兼容老版本系统）
FROM ubuntu:18.04

# 安装必要的工具和 CGO 依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    wget \
    libusb-1.0-0-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装 Go 1.23
RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz \
    && rm go1.23.4.linux-amd64.tar.gz

# 设置 Go 环境
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV GOOS="linux"
ENV GOARCH="amd64"
ENV CGO_ENABLED="1"

# 设置工作目录
WORKDIR /workspace

# 复制源码
COPY . /workspace/

# 下载依赖
RUN go mod download

# 编译（CGO 启用，支持串口和 HID）
RUN cd cmd && \
    go build -ldflags="-s -w" -o /workspace/hardware-test main.go

# 输出二进制文件
CMD ["cp", "/workspace/hardware-test", "/build/"]
