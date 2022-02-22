# post-requester

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/natron-io/post-requester)](https://goreportcard.com/report/github.com/natron-io/post-requester)

A microservice application, which gets files from a SMB share and post their content to an API Endpoint.

## Env Variables
Set the following required Env Variables to start the HTTP Post Handler
```bash
export ENDPOINT_USERNAME=<your username>
export ENDPOINT_PASSWORD=<your password>
export ENDPOINT_URL=<api endpoint url>
export SMB_SERVERNAME=<IP or DNS> # 192.168.1.10
export SMB_SHARENAME=<sharename> # The name of the Windows Share (not \\192.168.1.10\share, only share)
export SMB_USERNAME=<smb username> # without domain
export SMB_PASSWORD=<smb password>
export SMB_DOMAIN=<windows domain> # e.g. domain.local
```
**optional** you can set another server port with the following Env Variable (default is Port 80)
```bash
export INTERVAL_SECONDS=<seconds> # Default set to 60
```

## Docker

This docker run command deploys the post-handler without the smb access:
```bash
docker run -d -e ENDPOINT_USERNAME=<your username> \
-e ENDPOINT_PASSWORD=<your password> \
-e ENDPOINT_URL=<api endpoint> \
-e SMB_SERVERNAME=<IP or DNS> \
-e SMB_SHARENAME=<sharename> \
-e SMB_USERNAME=<smb username> \
-e SMB_PASSWORD=<smb password> \
-e SMB_DOMAIN=<windows domain> \
dockerbluestone/post-requester:latest
```

### Docker-Compose
```yaml
version: "3.9"  # optional since v1.27.0
services:
  post-requester:
    image: ghcr.io/natron.io/post-requester:latest
    environment:
      - ENDPOINT_USERNAME=username
      - ENDPOINT_PASSWORD=password
      - ENDPOINT_URL=https://api.test.com/upload
      - SMB_ENABLED=true
      - SMB_SERVERNAME=192.168.1.10
      - SMB_SHARENAME=share
      - SMB_USERNAME=username
      - SMB_PASSWORD=password
      - SMB_DOMAIN=domain.local
    networks: 
      - post-requester

networks:
  post-requester:
    driver: bridge
```

