项目说明
===

##  

## 隧道建立操作

```bash
# 准备动作
mkdir -p ~/.ssh && touch ~/.ssh/config
# 私钥
cat <<'EOF'>~/.ssh/jenkins@moyi-lc.com
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACATQwMkiaCP7I4nOFvBPomMZJT+1PTyLnkwyHuaXJWDMQAAAKArmOrvK5jq
7wAAAAtzc2gtZWQyNTUxOQAAACATQwMkiaCP7I4nOFvBPomMZJT+1PTyLnkwyHuaXJWDMQ
AAAEAfpUur6iMOKhzxRc4TiWkKCBXLLSuY3JU+V4PLyiYGrxNDAySJoI/sjic4W8E+iYxk
lP7U9PIueTDIe5pclYMxAAAAFnlvdXJfZW1haWxAZXhhbXBsZS5jb20BAgMEBQYH
-----END OPENSSH PRIVATE KEY-----
EOF

# 设置认证信息
sed -i -e '/# moyi-lc.com start/,/# moyi-lc.com end/d' ~/.ssh/config
cat <<'EOF'>>~/.ssh/config
# moyi-lc.com start
Host moyi-lc.com
  User jenkins
  Port 49183
  IdentityFile ~/.ssh/jenkins@moyi-lc.com
# moyi-lc.com end
EOF
chmod 600 ~/.ssh/config ~/.ssh/jenkins@moyi-lc.com


# 建立隧道

```

## FAQ

- [golang工程最佳实践](https://github.com/golang-standards/project-layout)