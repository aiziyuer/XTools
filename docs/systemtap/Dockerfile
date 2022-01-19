ARG VERSION=7.6.1810
ARG KERNEL_VERSION=3.10.0-957.21.3.el7

FROM centos:7.9.2009
ARG VERSION
ARG KERNEL_VERSION

ENV VERSION=$VERSION
ENV KERNEL_VERSION=$KERNEL_VERSION

# 配置私人加速源
RUN true \
&& sed -i 's%^mirrorlist=%#mirrorlist=%g; s%^#baseurl=%baseurl=%g' /etc/yum.repos.d/*.repo \
&& sed -i 's%http://debuginfo.centos.org/%http://mirrors.moyi-lc.com:5000/centos-debuginfo/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://vault.centos.org/%http://mirrors.moyi-lc.com:5000/centos-vault/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://mirror.centos.org/altarch%http://mirrors.moyi-lc.com:5000/centos-altarch/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://mirror.centos.org/%http://mirrors.moyi-lc.com:5000/%g' /etc/yum.repos.d/CentOS-*.repo \
&& true

# 安装软件包
RUN \
  yum install -y \
        systemtap systemtap-runtime bash-completion \
  && yum clean all

# 必要的软件包
RUN true \
&& yum --enablerepo=C${VERSION}-* \
    install -y kernel-${KERNEL_VERSION} kernel-devel-${KERNEL_VERSION} \
&& yum clean all

RUN \
  yum --enablerepo base-debuginfo install -y \
        kernel-debuginfo-${KERNEL_VERSION} \
  && yum clean all

# 测试生效
#stap -ve 'probe begin { log("hello world") exit() }'

