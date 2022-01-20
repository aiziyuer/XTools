ARG VERSION=7.6.1810
# ARG KERNEL_VERSION=3.10.0-957

FROM centos:7.9.2009
ARG VERSION
# ARG KERNEL_VERSION

ENV VERSION=$VERSION
# ENV KERNEL_VERSION=$KERNEL_VERSION

# 配置私人加速源
RUN true \
&& sed -i 's%^mirrorlist=%#mirrorlist=%g; s%^#baseurl=%baseurl=%g' /etc/yum.repos.d/*.repo \
&& sed -i 's%http://debuginfo.centos.org/%http://mirrors.moyi-lc.com:5000/centos-debuginfo/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://vault.centos.org/%http://mirrors.moyi-lc.com:5000/centos-vault/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://mirror.centos.org/altarch/%http://mirrors.moyi-lc.com:5000/centos-altarch/%g' /etc/yum.repos.d/CentOS-*.repo \
&& sed -i 's%http://mirror.centos.org/%http://mirrors.moyi-lc.com:5000/%g' /etc/yum.repos.d/CentOS-*.repo \
&& true

# 安装软件包
RUN \
  yum install -y \
        systemtap systemtap-runtime bash-completion \
  && yum clean all

# 尽可能使能必要的仓库
RUN true \
&& yum-config-manager --disable {base,updates} \
&& yum-config-manager --enable C${VERSION}-* \
&& yum repolist | grep C${VERSION} || yum-config-manager --enable {base,updates} \
&& yum-config-manager --enable base-debuginfo || true \
&& true

# 必要的软件包: kernel, kernel-devel, kernel-debuginfo
# RUN true \
# && yum --showduplicates list kernel | awk '{print $2}' \
#     | xargs -n1 -I{} yum install -y kernel-{} kernel-devel-{} kernel-debuginfo-{} kernel-debuginfo-common-*-{}\
# && yum clean all

# 恢复仓库
RUN true \
&& yum-config-manager --enable {base,updates} \
&& true


# 测试生效
#stap -ve 'probe begin { log("hello world") exit() }'

