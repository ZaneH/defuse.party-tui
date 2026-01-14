# defuse.party

A Terminal User Interface (TUI) recreation of a popular bomb defusal game served over SSH.

## Overview

This project provides a text-based interface for playing defuse.party, connectable via SSH. It connects to the existing
[Go gRPC backend](https://github.com/ZaneH/keep-talking) for game logic.

## Quick Start

### Prerequisites

- Go 1.21+
- The gRPC backend server running

**Note**: This TUI imports protocol buffer definitions from the [backend repository](https://github.com/ZaneH/keep-talking) as a Go module dependency. Proto files are not duplicated in this repository - they are consumed directly from the backend's generated code.

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

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TUI_SSH_PORT` | `2222` | SSH listen port |
| `TUI_GRPC_ADDR` | `localhost:50051` | gRPC backend address |
