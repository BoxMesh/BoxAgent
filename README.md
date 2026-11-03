
# EdgeAgent

[![Build](https://github.com/DeviceTunnel/EdgeAgent/actions/workflows/build.yml/badge.svg)](https://github.com/DeviceTunnel/EdgeAgent/actions/workflows/build.yml)

**EdgeAgent** is a lightweight, cross-platform background agent built with C# and .NET. It runs as a system service on Windows and Linux machines, enabling remote management, system monitoring, and secure communication with EdgePanel.

Each machine runs a single EdgeAgent instance that maintains a persistent outbound gRPC connection to EdgePanel. This allows remote communication without requiring a public IP address or inbound firewall ports.

## Features

- **System Monitoring** – Collects CPU, memory, disk, network, and system information.
- **Remote Command Execution** – Receives and executes authorized commands.
- **Reverse Proxy** – Provides access to local applications through a remote tunnel.
- **File Management** – Supports remote file operations.
- **Persistent Connection** – Maintains a bidirectional gRPC connection with automatic reconnection.
- **Secure Communication** – Uses TLS/mTLS for authenticated and encrypted communication.
- **Cross-platform** – Supports Windows and Linux.

## Architecture

```text
Windows / Linux Machine
          │
          ▼
     EdgeAgent
          │
     ┌────┴──────────────┐
     │                   │
  Monitoring          Remote Access
     │                   │
  CPU / RAM          HTTP Proxy
  Disk / Network     Command
                     File Management
          │
          ▼
   gRPC Bidirectional
        Streaming
          │
         mTLS
          │
          ▼
       EdgePanel
```

EdgeAgent is designed as a single worker service with modular capabilities. All features share the same agent lifecycle and communication infrastructure.

## Technology Stack

| Component         | Technology                   |
| ----------------- | ---------------------------- |
| Language          | C#                           |
| Runtime           | .NET                         |
| Application Model | Worker Service               |
| Communication     | gRPC Bidirectional Streaming |
| Security          | TLS / mTLS                   |
| Platforms         | Windows, Linux               |

## Communication

EdgeAgent establishes an outbound connection to EdgePanel and keeps it alive for bidirectional communication.

- Agent initiates the connection.
- EdgePanel can send commands and configuration updates.
- Agent reports system metrics and health information.
- Multiple operations can be processed concurrently.
- Connection recovery is handled automatically.

## Requirements

- .NET Runtime (if using framework-dependent deployment).
- Windows or Linux operating system.
- Network access to the EdgePanel endpoint.
- Valid agent credentials and certificates.

## Installation

Installation instructions will be provided as platform-specific deployment packages and scripts.

## Development

Clone the repository:

```bash
git clone https://github.com/DeviceTunnel/EdgeAgent.git
cd EdgeAgent
```

Build:

```bash
dotnet build
```

Run:

```bash
dotnet run
```

## Project Status

EdgeAgent is under active development. Features and communication protocols may evolve as the platform grows.

## License

To be determined.
