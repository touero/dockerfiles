# gitweb

A GitWeb + nginx container image for quickly browsing Git repositories through a web UI.

## Highlights

- GitWeb integrated with nginx/fcgiwrap startup flow
- Suitable for read-only repository browsing in internal networks
- Supports multi-architecture image builds

## Usage

```bash
docker run --rm -p 8080:80 -v /path/to/repos:/var/lib/git weiensong/git:latest
```
