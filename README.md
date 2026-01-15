# defuse.party: tui

A Terminal User Interface (TUI) recreation of a popular bomb defusal game served over SSH.

## Overview

This project provides a text-based interface for playing defuse.party, connectable via SSH. It connects to the existing
[Go gRPC backend](https://github.com/ZaneH/defuse.party-go) for game logic.

## Quick Start

### Prerequisites

- Go 1.21+
- The gRPC backend server running

**Note**: This TUI imports protocol buffer definitions from the [backend repository](https://github.com/ZaneH/defuse.party-go) as a Go module dependency. Proto files are not duplicated in this repository - they are consumed directly from the backend's `pkg/proto` package.

### Building

```bash
make build
```

This creates the `tui-server` binary.

### Running

```bash
# Set the gRPC server address if not localhost:50051
export TUI_GRPC_ADDR=localhost:50051

# Run the server (default SSH port: 2222)
./tui-server
```

Or use the Makefile:
```bash
TUI_GRPC_ADDR=localhost:50051 make run
```

### Connecting

```bash
ssh -p 2222 localhost
# Or from another machine:
ssh -p 2222 <server-ip>
```

On first run, SSH host keys will be generated in `.ssh/`.

## Docker

### Building

```bash
docker build -t defuse-party:latest .
```

The Dockerfile uses `go get` to fetch the backend proto package from GitHub. Ensure the backend changes are committed and pushed before building.

### Running

```bash
docker run -p 2222:2222 -e TUI_GRPC_ADDR=host.docker.internal:50051 defuse-party:latest
```

## Development Workflow

### Production Mode (Default)

By default, the TUI fetches proto definitions from the published backend repository on GitHub:

```bash
# Update to latest backend version
go get -u github.com/ZaneH/defuse.party-go@latest
go mod tidy
```

This mode is used for:
- Docker builds
- Production environments

### Local Development Mode

For rapid local development with proto changes, you can use the local backend:

1. Edit `go.mod` and uncomment the `replace` directive:
   ```go
   replace github.com/ZaneH/defuse.party-go => ../keep-talking
   ```

2. Make changes to proto files in `../keep-talking/proto/`

3. Regenerate proto files in the backend:
   ```bash
   cd ../keep-talking
   buf generate
   ```

4. Test locally (TUI will use your local backend changes)

5. When ready to deploy:
   - Commit and push backend changes
   - Re-comment the `replace` directive in `go.mod`
   - Run `go get -u github.com/ZaneH/defuse.party-go@latest`
   - Commit TUI with updated `go.mod`

**Important**: Never commit with the `replace` directive uncommented - it breaks production builds.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TUI_SSH_PORT` | `2222` | SSH listen port |
| `TUI_GRPC_ADDR` | `localhost:50051` | gRPC backend address |
