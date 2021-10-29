ARG VERSION=7.6.1810
ARG KERNEL_VERSION=3.10.0-1160.42.2.el7


FROM centos:$VERSION
ARG VERSION
ARG KERNEL_VERSION

ENV VERSION=$VERSION
ENV KERNEL_VERSION=$KERNEL_VERSION

RUN \
rm -rf /etc/yum.repos.d/*.repo \
&& printf "\n\
[base]\n\
name=Base\n\
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/${VERSION}/os/\$basearch/\n\
gpgcheck=0\n\
enabled=1\n\
\n\
[updates]\n\
name=Updates\n\
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/${VERSION}/updates/\$basearch/\n\
gpgcheck=0\n\
enabled=1\n\
\n\
[extras]\n\
name=Extras\n\
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/${VERSION}/extras/\$basearch/\n\
gpgcheck=0\n\
enabled=1\n\
\n\
[centosplus]\n\
name=CentOSPlus\n\
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/${VERSION}/centosplus/\$basearch/\n\
gpgcheck=0\n\
enabled=1\n\
\n\
[fasttrack]\n\
name=Fasttrack\n\
baseurl=https://mirrors.tuna.tsinghua.edu.cn/centos-vault/${VERSION}/fasttrack/\$basearch/\n\
gpgcheck=0\n\
enabled=1\n\
\n\
[base-debuginfo]\n\
name=Debuginfo\n\
baseurl=http://debuginfo.centos.org/7/\$basearch/\n\
gpgcheck=0\n\
enabled=0\n\
" > /etc/yum.repos.d/CentOS.repo

# 安装软件包
RUN \
  yum install -y \
        systemtap systemtap-runtime bash-completion \
  && yum clean all

# 必要的软件包
RUN \
  yum install -y \
        kernel-${KERNEL_VERSION} kernel-devel-${KERNEL_VERSION} \
  && yum clean all

RUN \
  yum --enablerepo base-debuginfo install -y \
        kernel-debuginfo-${KERNEL_VERSION} kernel-debuginfo-common-$(uname -m)-${KERNEL_VERSION} \
  && yum clean all

# 测试生效
#stap -ve 'probe begin { log("hello world") exit() }'

