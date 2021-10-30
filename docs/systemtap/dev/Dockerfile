FROM centos:7


RUN \
  yum install -y \
        systemtap systemtap-runtime bash-completion \
  && yum clean all

# 必要的软件包
RUN \
  yum install -y \
        kernel-$(uname -r) kernel-devel-$(uname -r) \
  && yum clean all

RUN \
  yum --enablerepo base-debuginfo install -y \
        kernel-debuginfo-$(uname -r) kernel-debuginfo-common-$(uname -m) \
  && yum clean all

RUN \
  yum install centos-release-scl -y \
  && yum install devtoolset-9-* -y

RUN \
  yum install git 


