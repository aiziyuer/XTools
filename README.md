项目说明
===


## XTunnel使用

```bash
# 建立隧道
./XTunnel run

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