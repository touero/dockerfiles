# fileserver

A lightweight static file server image based on Python, intended for personal and small-team LAN distribution scenarios.

## Highlights

- Simple Python HTTP service startup
- Minimal image footprint
- Easy volume mounting for directory sharing

## Usage

```bash
docker run --rm -p 8000:8000 -v "$PWD:/data" weiensong/fileserver:latest
```
