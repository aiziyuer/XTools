FROM centos:7

ARG KERNEL_VERSION=3.10.0-1160.42.2.el7.x86_64
ENV KERNEL_VERSION=${KERNEL_VERSION:-'3.10.0-1160.42.2.el7.x86_64'}

# 安装软件包
RUN \
  yum install -y \
        systemtap systemtap-runtime applydeltarpm

# 必要的软件包
RUN \
  yum install -y \
        kernel-${KERNEL_VERSION} kernel-devel-${KERNEL_VERSION}

RUN \
  yum --enablerepo base-debuginfo install -y \
        kernel-debuginfo-${KERNEL_VERSION} kernel-debuginfo-common-$(uname -m)-${KERNEL_VERSION}

# 测试生效
#stap -ve 'probe begin { log("hello world") exit() }'

