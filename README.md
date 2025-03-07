# Portl

Portl is a lightweight SSH/SFTP server that provides a secure way to transfer files through an HTTP interface. It acts as a bridge between SSH/SFTP clients and HTTP endpoints, making it easy to integrate file transfer capabilities into web applications.

## Features

- SSH/SFTP server support
- HTTP interface for file operations
- Secure file transfer capabilities
- JSON-structured logging
- Configurable log levels

## Prerequisites

- Go 1.23 or later
- SSH private key for server authentication

## Installation

1. Clone the repository:
```bash
git clone https://github.com/danutavadanei/portl.git
cd portl
```

2. Build the project:
```bash
go build -o portl ./cmd/main.go
```

## Usage

1. Generate an SSH private key if you don't have one:
```bash
ssh-keygen -t rsa -b 4096 -f keys/id_rsa
```

2. Run the server:
```bash
./portl
```

Optional flags:
- `-debug`: Enable debug level logging

## Demo Server

A demo server is available at `portl.znd.ro`. You can use it to test file transfers:

### Single File Transfer
```bash
scp /path/to/my/file portl.znd.ro:/
```

### Directory Transfer
```bash
scp -r /path/to/my/directory portl.znd.ro:/
```

## Configuration

The server can be configured through environment variables or a config file:

- `SSH_LISTEN_ADDR`: SSH server listen address (default: ":2222")
- `HTTP_LISTEN_ADDR`: HTTP server listen address (default: ":8080")
- `HTTP_BASE_URL`: Base URL for HTTP endpoints
- `SSH_PRIVATE_KEY_PATH`: Path to SSH private key file

## License

This project is licensed under the MIT License - see the LICENSE file for details.