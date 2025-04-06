# Portl

*portl* is a lightweight peer-to-peer file transfer tool written in Go, designed for secure and efficient file sharing without persistent storage.

Using `scp` as the uploader and an HTTP client as the downloader, portl creates a seamless tunnel between two peers through a central server. Files are streamed directly from the uploader to the downloader and delivered as a .zip archive — ensuring minimal resource usage and no intermediate storage.

## Key Features
- Fast and secure file transfer
- No files stored on the server
- Files delivered as a .zip archive
- Uses scp and HTTP for simplicity
- Ideal for direct peer-to-peer sharing via a relay server

## How It Works
1. Uploader sends files via scp to the server.
2. Server tunnels the data to the downloader without storing it.
3. Downloader receives a zipped stream over HTTP.

## Demo Server

A demo server is available at `portl.znd.ro`. You can use it for sharing files

### Single File Transfer
```bash
scp /path/to/my/file portl.znd.ro:/
```

### Directory Transfer
```bash
scp -r /path/to/my/directory portl.znd.ro:/
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/danutavadanei/portl.git
cd portl
```

2. Build the project:
```bash
make build
```

## Usage

1. Generate an SSH private key if you don't have one:
```bash
ssh-keygen -t rsa -b 4096 -f keys/id_rsa
```

2. Run the server:
```bash
./bin/portl
```


## Configuration

The server can be configured through environment variables or a config file:

- `SSH_LISTEN_ADDR`: SSH server listen address (default: ":2222")
- `HTTP_LISTEN_ADDR`: HTTP server listen address (default: ":8080")
- `HTTP_BASE_URL`: Base URL for HTTP endpoints
- `SSH_PRIVATE_KEY_PATH`: Path to SSH private key file
- `DEBUG`: Enable debug logging (default: false)

## License

This project is licensed under the MIT License - see the LICENSE file for details.