
# 这个版本不支持, 因为2009这个版本没有进vault
# docker build --build-arg VERSION=7.9.2009 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7 . -t aiziyuer/centos7-systemtap:3.10.0-1160.42.2

docker build -f Dockerfile --build-arg VERSION=7.8.2003 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7 . 
docker build -f Dockerfile --build-arg VERSION=7.7.1908 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7 .

# 常用版本
docker build -f Dockerfile --build-arg VERSION=7.6.1810 --build-arg KERNEL_VERSION=3.10.0-957.21.3.el7 . -t aiziyuer/centos7-systemtap:3.10.0-957.21.3.el7

