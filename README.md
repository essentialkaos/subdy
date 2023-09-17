<p align="center"><a href="#readme"><img src="https://gh.kaos.st/subdy.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/r/subdy"><img src="https://kaos.sh/r/subdy.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/l/subdy"><img src="https://kaos.sh/l/395df260d03eaa6c8a31.svg" alt="Code Climate Maintainability" /></a>
  <a href="https://kaos.sh/b/subdy"><img src="https://kaos.sh/b/ad36c313-9009-4abe-97d6-7c1f0de39794.svg" alt="Codebeat badge" /></a>
  <a href="https://kaos.sh/w/subdy/ci"><img src="https://kaos.sh/w/subdy/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/subdy/codeql"><img src="https://kaos.sh/w/subdy/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#screenshots">Screenshots</a> • <a href="#installation">Installation</a> • <a href="#usage">Usage</a> • <a href="#ci-status">CI Status</a> • <a href="#contributing">Contributing</a> • <a href="#license">License</a></p>

<br/>

`subdy` is simple CLI for [subdomain.center](https://www.subdomain.center) API.

### Screenshots

<p align="center">
  <img src="https://gh.kaos.st/subdy.png" alt="subdy preview">
</p>

### Installation

#### From source

To build the `subdy` from scratch, make sure you have a working Go 1.20+ workspace (_[instructions](https://go.dev/doc/install)_), then:

```bash
go install github.com/essentialkaos/subdy@latest
```

#### Prebuilt binaries

You can download prebuilt binaries for Linux and macOS from [EK Apps Repository](https://apps.kaos.st/subdy/latest):

```bash
bash <(curl -fsSL https://apps.kaos.st/get) subdy
```

#### Container Image

The latest version of `subdy` also available as container image on [GitHub Container Registry](https://kaos.sh/p/subdy) and [Docker Hub](https://kaos.sh/d/subdy):

```bash
podman run --rm -it ghcr.io/essentialkaos/subdy:latest mydomain.com
# or
docker run --rm -it ghcr.io/essentialkaos/subdy:latest mydomain.com
```

### Command-line completion

You can generate completion for `bash`, `zsh` or `fish` shell.

Bash:
```bash
sudo subdy --completion=bash 1> /etc/bash_completion.d/subdy
```

ZSH:
```bash
sudo subdy --completion=zsh 1> /usr/share/zsh/site-functions/subdy
```

Fish:
```bash
sudo subdy --completion=fish 1> /usr/share/fish/vendor_completions.d/subdy.fish
```

### Man documentation

You can generate man page using next command:

```bash
subdy --generate-man | sudo gzip > /usr/share/man/man1/subdy.1.gz
```

### Usage

```
Usage: subdy {options} domain

Options

  --ip, -I                 Resolve subdomains IP
  --dns, -D name-or-url    DoH provider (cloudflare|google|url)
  --no-color, -nc          Disable colors in output
  --help, -h               Show this help message
  --version, -v            Show version

Examples

  subdy go.dev
  Find all subdomains of go.dev

  subdy -I go.dev
  Find all subdomains of go.dev and resolve their IPs

  subdy -I -D google go.dev
  Find all subdomains of go.dev and resolve their IPs using Google DNS
```

### CI Status

| Branch | Status |
|--------|----------|
| `master` | [![CI](https://kaos.sh/w/subdy/ci.svg?branch=master)](https://kaos.sh/w/subdy/ci?query=branch:master) |
| `develop` | [![CI](https://kaos.sh/w/subdy/ci.svg?branch=develop)](https://kaos.sh/w/subdy/ci?query=branch:develop) |

### Contributing

Before contributing to this project please read our [Contributing Guidelines](https://github.com/essentialkaos/contributing-guidelines#contributing-guidelines).

### License

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
