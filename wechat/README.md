# wechat

A containerized WeChat desktop runtime environment adapted for Linux container workflows. Contains fcitx5 Chinese input method and novnc.

## Highlights

- Encapsulated startup scripts for easier runtime orchestration
- Designed for amd64 and arm64 builds
- Can be combined with X11/VNC related runtime environments

## Usage

### deploy
```bash
docker run -d --name wechat -it weiensong/wechat:latest
```

### test
```bash
docker run --rm -it weiensong/wechat:latest
```


### visit
```
http://127.0.0.1:6080/?autoconnect=1&password=secret&resize=scale&null
```
