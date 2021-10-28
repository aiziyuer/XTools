

docker build --build-arg VERSION=7.9.2009 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7.x86_64 .
docker build --build-arg VERSION=7.8.2003 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7.x86_64 .
docker build --build-arg VERSION=7.7.1908 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7.x86_64 .

# 常用版本
docker build --build-arg VERSION=7.6.1810 --build-arg KERNEL_VERSION=3.10.0-957.21.3.el7.x86_64 . -t aiziyuer/centos7-systemtap:3.10.0-957.21.3.el7.x86_64

