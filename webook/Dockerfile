# 基于 Ubuntu 20.04 创建一个镜像
FROM ubuntu:20.04
# 将本地的 webook 文件复制到容器的 /app 目录下
COPY webook /app/webook
# 设置容器的工作目录为 /app
WORKDIR /app
# 指定容器启动后直接运行 /app/webook(上一步设置了工作目录)
ENTRYPOINT ["/app/webook"]