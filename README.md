项目说明
===


## XTunnel使用

```bash
# 快速建立隧道
./XTunnel run -ssh_tunnels 'R=>0.0.0.0:3128=>localhost:3128' --ssh_uri root@aliyun.moyi-lc.com:22 --ssh_password XXXX

# 批量建立隧道
cat <<EOF>~/.config/XTunnel/XTunnel.yaml
---
# eg. [ssh_user]@[ssh_host]:[ssh_port]
ssh_uri: root@127.0.0.1:22
ssh_password:
ssh_identity:
ssh_proxy_uri: 
ssh_tunnels:
# 监听端口: nc -lp 4444  查看端口: netstat -anp
# R: 反向隧道, L: 正向隧道, D: ss代理, H: 原地http代理, S:原地ss代理
# - R=>0.0.0.0:13141=>localhost:3389
# - L=>0.0.0.0:12345=>localhost:22
# - LD=>0.0.0.0:11080
# - LH=>0.0.0.0:3128
# - RD=>0.0.0.0:11080
# - RH=>0.0.0.0:3128
EOF
./XTunnel server

```

## NVIDIAMate使用

```
# 启动
mkdir -p ~/.config/NVIDIAMate
cp config/NVIDIAMate.yaml ~/.config/NVIDIAMate/
go run app/cmd/NVIDIAMate run

# 添加systemd服务
mkdir -p /var/log/NVIDIAMate
cp init/NVIDIAMate.service /etc/systemd/system/NVIDIAMate.service
systemctl enabel NVIDIAMate
systemctl start NVIDIAMate
systemctl status NVIDIAMate
```


## FAQ

- [golang工程最佳实践](https://github.com/golang-standards/project-layout)
- [yaml与json互转](https://github.com/ghodss/yaml)
- []