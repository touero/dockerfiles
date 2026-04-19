# gitweb

A GitWeb + nginx container image for quickly browsing Git repositories through a web UI.Use [gitweb-dark-theme](https://github.com/CBenoit/gitweb-dark-theme) theme by default


## Highlights

- GitWeb integrated with nginx/fcgiwrap startup flow
- Suitable for read-only repository browsing in internal networks
- Custom dark theme
- Supports multi-architecture image builds


> - linux/386
> - linux/amd64
> - linux/arm/v6
> - linux/arm/v7
> - linux/arm64
> - linux/ppc64le
> - linux/s390x

## Usage

### deploy

```bash
docker run -d --name gitweb --restart unless-stopped -v /path/bare_repos:/srv/git:ro -p 80:80 weiensong/gitweb
```

### dev

```bash
docker run --name gitweb /path/bare_repos:/srv/git:ro -p 80:80 weiensong/gitweb
```

