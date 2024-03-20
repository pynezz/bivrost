# Bivrost / Bifrost

> In Norse mythology, Bifröst (/ˈbɪvrɒst/ [ⓘ] [[1]]), also called Bilröst, is a burning rainbow bridge that reaches between Midgard (Earth) and Asgard, the realm of the gods.
> [Wikipedia](https://en.wikipedia.org/wiki/Bifröst)

[1]: https://www.collinsdictionary.com/dictionary/english/bifrost?showCookiePolicy=true
[ⓘ]: https://en.wikipedia.org/wiki/File:Bifrost.ogg

---
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)

<img src="./img/bivrost_readme_cover.png" width="100%" alt="Bivrost cover photo"/>

## Overview

Bivrost is a simple, (hopefully) fast, and (hopefully) reliable adapter and bridge between different services. It's designed to be modular and easy to extend, and to be able to handle a wide variety of different services and protocols.

Bivrost is written in Go due to it being a statically typed, memory safe, and compiled language designed for networking and concurrency.

## Purpose

Bivrost serves as a log normalization and aggregation service, which is designed to be able to handle a wide variety of different services and protocols. It is designed to be modular and easy to extend, and to be able to handle a wide variety of different services and protocols.

## Configuration

Bivrost is configured using a simple configuration file, which is written in YAML. The configuration file is used to specify the services and protocols that Bivrost should handle, as well as the settings for each service and protocol.

### Configuration Example

This will have to be revised as the project progresses.

```json
{
    "sources": [
        {
            "name": "siem logs",
            "type": "directory",
            "location": "/var/log/siem",
            "format": "json",
            "tags": ["siem", "logs"]
        },
        {
            "name": "syslog",
            "type": "service",
            "location": " ",
            "format": "json",
            "tags": ["syslog", "logs"]
        },
        {
            "name": "threat intel",
            "type": "module",
            "location": "/path/to/module/output",
            "format": "json",
            "tags": ["intel", "module"]
        },
        {
            "name": "thevalve",
            "type": "module",
            "location": " ",
            "format": "raw",
            "tags": ["credentials", "module"]
        },
        {
            "name": "<user-defined-name>",
            "type": "<'module', 'logtype', 'directory', 'service'>",
            "location": " ",
            "format": "<'json', 'raw', 'ascii'>",
            "tags": ["credentials", "module"]
        }
    ],
    "<some-other-key?>": "<?>"
}
```

We might need to evaluate if YAML simply is better for our purpose, considering the config file will be used by users that might find this format more intuitive.
> We decided to go with YAML 👍

```yaml
# YAML also supports comments, which means we can guide the user through the configuration file.
sources:
  - name: siem logs
    type: directory
    location: /var/log/siem
    format: json
    tags:
      - siem
      - logs
# Here ends the first source
# Here starts a new source.
  - name: syslog    # Arbitrary name of the source
    type: service   # Type of the source
    location: ' '   # Location of the source(such as path, uri, rpc, and ipc)
    format: json    # Format bivfrost should expect
    tags:           # Tags to be used for filtering
      - syslog      # Tag for filtering
      - logs        # Tag for filtering

  - name: threat intel
    type: module
    location: /path/or/uri/to/module/output
    format: json
    tags:
      - intel
      - module
  - name: thevalve
    type: module
    location: ' '
    format: raw
    tags:
      - credentials
      - module
  - name: <user-defined-name>
    type: <module/logtype/directory/service>
    location: ' '
    format: <json, raw, ascii>
    tags:
        - <descriptive tag>
        - ...no tag limit

network:
  - read_timeout: 10  # time in seconds
  - write_timeout: 10 # time in seconds'

users_database:
  path: /path/to/users.db

```

## Usage

Bivrost is designed to be easy to use and to require minimal configuration. It is designed to be self-contained and to require no dependencies.

```bash
bivrost --config /path/to/config.yaml
```

### Help Output

```bash
$ bivrost -h

Usage:
  --config <string>     Path to the configuration file (default "config.yaml")
  --version             Print version information

  -h, --help            Print this help message
```

## Features

- **Self-contained**: Bivrost is a single binary with no dependencies.
- **Modular**: Bivrost is designed to be easy to extend and add new services and protocols.
- **Fast**: Bivrost is designed to be fast and efficient.
- **Reliable**: Bivrost is designed to be reliable and to handle a wide variety of different services and protocols.

## Integrated with TheValve

Bivrost is integrated with [TheValve](https://github.com/pynezz/thevalve), where TheValve serves as a secure storage and cryptographic service for Bivrost.

## Authentication

- Hashing algorithm and key derivation: Argon2 HMAC-SHA256
- Tokens (JWT-HS256)

## Database

- sqlite3

## Packages

- [Go Fiber](https://gofiber.io/)
- [go-sqlite3](https://github.com/mattn/go-sqlite3)

## Requirements for compiling

- Go version > 1.21
- gcc *(for go-sqlite3 as it requires cgo)*

### Protobuf

- [Protocol Buffers](https://developers.google.com/protocol-buffers)

```bash
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```


## License

*TBD*
