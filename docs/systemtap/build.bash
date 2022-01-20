
# 这个版本不支持, 因为2009这个版本没有进vault
# docker build --build-arg VERSION=7.9.2009 --build-arg KERNEL_VERSION=3.10.0-1160.42.2.el7 . -t aiziyuer/centos7-systemtap:3.10.0-1160.42.2

docker build --build-arg VERSION=7.9.2009 --build-arg KERNEL_VERSION=3.10.0-1160 . -t aiziyuer/centos7-systemtap:3.10.0-1160
docker build --build-arg VERSION=7.8.2003 --build-arg KERNEL_VERSION=3.10.0-1127 . -t aiziyuer/centos7-systemtap:3.10.0-1127
docker build --build-arg VERSION=7.7.1908 --build-arg KERNEL_VERSION=3.10.0-1062 . -t aiziyuer/centos7-systemtap:3.10.0-1062
docker build --build-arg VERSION=7.6.1810 --build-arg KERNEL_VERSION=3.10.0-957 . -t aiziyuer/centos7-systemtap:3.10.0-957
docker build --build-arg VERSION=7.5.1804 --build-arg KERNEL_VERSION=3.10.0-862 . -t aiziyuer/centos7-systemtap:3.10.0-862
docker build --build-arg VERSION=7.4.1708 --build-arg KERNEL_VERSION=3.10.0-693 . -t aiziyuer/centos7-systemtap:3.10.0-693
docker build --build-arg VERSION=7.3.1611 --build-arg KERNEL_VERSION=3.10.0-514 . -t aiziyuer/centos7-systemtap:3.10.0-514
docker build --build-arg VERSION=7.2.1511 --build-arg KERNEL_VERSION=3.10.0-327 . -t aiziyuer/centos7-systemtap:3.10.0-327
docker build --build-arg VERSION=7.1.1503 --build-arg KERNEL_VERSION=3.10.0-229 . -t aiziyuer/centos7-systemtap:3.10.0-229
docker build --build-arg VERSION=7.0.1406 --build-arg KERNEL_VERSION=3.10.0-123 . -t aiziyuer/centos7-systemtap:3.10.0-123

