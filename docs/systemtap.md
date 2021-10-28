SystemTap学习
===


## 运行时环境-CentOS

``` 
# 必要的软件包
yum install -y systemtap systemtap-runtime
yum install -y kernel-$(uname -r) kernel-devel-$(uname -r)
yum --enablerepo base-debuginfo install -y kernel-debuginfo-$(uname -r)
yum --enablerepo base-debuginfo install -y kernel-debuginfo-common-$(uname -m)-$(uname -r)

# 测试生效
stap -ve 'probe begin { log("hello world") exit() }'
```

