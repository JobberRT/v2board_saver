# v2board_saver
A simple auto port changer for your v2board panel.

## What does it do?
It will auto change port when your node's port is blocked.

## Before use
This is only tested on my own v2board panel. Make a panel backup, Test it before use it in production! Any damage or problem that can not be recovered is not responsible to me!

## How to use?

### Docker ENV
- `TIMEOUT`: timeout that used in TCP check port and HTTP request
- `EMAIL`: v2board admin login email
- `PASSWORD`: v2board admin login password
- `INTERVAL`: how much time between two check
- `URL`: v2board panel's url, like "https://vpn.example.com"
- `TZ`: set it to your own timezone, default is `Asia/Shanghai`

### Pre-Built Docker image
**AMD64**
`docker run jobbert/v2board_saver-amd64`
<br>
**ARM64**
`docker run jobbert/v2board_saver-arm64`

### Build your own Docker image
```shell
git clone git@github.com:JobberRT/v2board_saver.git
cd v2board_saver

# NOTICE: check the docker file and modify it if needed
# cat Dockerfile
# vim Dockerfile

docker build . -t example_name:example_version
```