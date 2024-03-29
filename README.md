![OrcaCD](img/orcacd_logo_text.svg)

[![Github release](https://img.shields.io/github/v/release/7oku/orcacd.svg)](https://github.com/7oku/orcacd/releases)
[![License](https://img.shields.io/github/license/7oku/orcacd.svg)](http://www.apache.org/licenses/LICENSE-2.0)

OrcaCD is a small tool to bring GitOps CD to docker-compose.yml files, similar to ArgoCD/FluxCD for kubernetes.

## Installation

Copy docker-compose.yml somewhere on your server

```yml
version: "3"

services:
  orcacd: 
    image: ghcr.io/7oku/orcacd:latest
    tty: true
    ports: 
      - "6688:6666"
    environment:
      # color fixes for logging
      CI: true
      TERM: xterm-256color
      COLORTERM: truecolor
      # example repo
      OCD_REPOS_EXAMPLE_URL: "https://raw.githubusercontent.com/7oku/orcacd/main/compose-files/servicename/docker-compose.yaml"
      OCD_REPOS_EXAMPLE_USER: "7oku"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    restart: unless-stopped
```

Fire up and check for logs:

```console
foo@bar:~$ docker compose up -d
foo@bar:~$ docker compose logs -fn200
```

## Configuration

### - via config.yml

1. add volume to *docker-compose.yml*

   ```yml
   volumes:
         - ./config.yml:/config.yml
   ```

2. create/copy config.yml

    ```yml
    repos:
      githubsuperservice:
        url: "https://raw.githubusercontent.com/7oku/orcacd/main/compose-files/servicename/docker-compose.yaml"
   basicauth:
     user1: "god"
   loglevel: "error"
   autosync: "on"
   targetpath: "/tmp/opt/compose"
   workdir: "/tmp/ocd"
   ```

### - via environment variables in docker-compose.yml prefixed with `OCD_`

| ENV VAR | default* / possible values | Description |
|---|---|---|
| OCD_REPOS_EXAMPLE1_URL *(required)* | (`string`) nil* | URL to compose file |
| OCD_REPOS_EXAMPLE1_USER *(optinoal)* | (`string`) nil* | User if required for auth |
| OCD_REPOS_EXAMPLE1_SECRET *(optional)* | (`string`) nil* | Either password or a Token (if used with github/gitlab) |
| OCD_LOGLEVEL *(optional)* | (`string`) debug,info,warning,error* | Log level for output |
| OCD_AUTOSYNC *(optional)* | (`string`) on*,off | Enable or disable autosync on startup |
| OCD_INTERVAL *(optional)* | (`int`) 300* | Interval to check for new file in seconds |
| OCD_WORKDIR *(optional)* | (`string`) /tmp/ocd* | Where ocd stores temp files |
| OCD_TARGETPATH *(optional)* | (`string`) /tmp/ocd/opt/compose* | Path to place resulting services |
| OCD_BASICAUTH_OCDADMIN *(optional)* | (`string`) ocd1337god* | BasicAuth for API (WIP) |
