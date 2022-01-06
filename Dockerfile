#### 第一阶段，用golang镜像编译
FROM golang:1.17.5-alpine3.15 AS builder

ENV TZ="Asia/Shanghai"

COPY . /opt/code

WORKDIR /opt/code

RUN set -ex \
    # apk 加速
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add tzdata --no-cache \
    # go 加速
    && go env -w GO111MODULE=on \
    && go env -w GOPROXY=https://goproxy.cn,direct \
    && go build -o confd

#### 第二阶段，用alpine镜像处理生成最终镜像
FROM alpine:3.15 AS prod

ENV TZ="Asia/Shanghai"

COPY --from=builder /opt/code/confd /usr/local/bin/confd 
COPY --from=builder /usr/share/zoneinfo/${TZ} /usr/share/zoneinfo/${TZ} 
COPY ./entrypoint.sh /

RUN set -ex \
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add vim bash gettext --no-cache \
    && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo "${TZ}" > /etc/timezone \
    && mkdir -p /etc/confd/templates /etc/confd/conf.d /tmp/acm \
    && chmod u+x /entrypoint.sh


COPY ./templates/constants.template /etc/confd/templates/constants.template
COPY ./templates/template.toml /etc/confd/conf.d/template.toml


ENTRYPOINT [ "/entrypoint.sh" ]
