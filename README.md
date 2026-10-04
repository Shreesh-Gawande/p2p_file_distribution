# Peer-to-Peer File Distribution

A small Go TCP application for sending a file from one computer to another. One node listens for incoming connections, and a sending node connects to it, transfers one file, then exits.

> **Security warning:** This project is intended for testing on a trusted local network only. It currently has no peer authentication and does not encrypt the network connection. Do not expose it to the public internet or send sensitive files with it.

## Requirements

- Windows computers for the included `p2p-share.exe` build. To build from source, install Go compatible with the version declared in `go.mod` (`go 1.25.5`).
- Both computers connected to a network that allows them to reach one another. For the simplest test, connect them to the same Wi-Fi or LAN.
- The receiving computer's IPv4 address and listening port.
- An inbound Windows Firewall rule allowing the listening TCP port on the receiving computer's **Private** network profile, if Windows blocks the connection.

## Build

Open PowerShell in the project directory (the directory containing `go.mod`):

```powershell
Set-Location "C:\path\to\p2p_file_distribution"
go test ./...
go build -o p2p-share.exe .
```

Copy `p2p-share.exe` to each Windows computer. Go and the source project are not required on computers that only run the executable. Build a separate executable for a different operating system or CPU architecture.

Show command-line help:

```powershell
.\p2p-share.exe -h
```

## Send a file on a local network

### 1. Start the receiver

On the computer that will receive the file, open PowerShell in the folder containing `p2p-share.exe` and start a listening node:

```powershell
.\p2p-share.exe -listen :3000
```

Leave this window open while the transfer takes place. The receiver listens on port `3000` on all network interfaces.

Find the receiver's IPv4 address by running:

```powershell
ipconfig
```

Use the IPv4 address for the Wi-Fi or Ethernet adapter connected to the same network. Do not use `127.0.0.1` when connecting from another computer.

### 2. Allow the connection if Windows Firewall blocks it

If Windows shows a firewall prompt, allow the application on the **Private** network. If there is no prompt and the sender cannot connect, an administrator can add a narrowly scoped inbound rule on the receiving computer from an elevated PowerShell window:

```powershell
New-NetFirewallRule `
  -DisplayName "P2P file test" `
  -Direction Inbound `
  -Protocol TCP `
  -LocalPort 3000 `
  -Action Allow `
  -Profile Private
```

Only allow this on a trusted network. Remove the rule after testing if it is no longer needed:

```powershell
Remove-NetFirewallRule -DisplayName "P2P file test"
```

### 3. Send the file

On the sending computer, open a second PowerShell window in the folder containing `p2p-share.exe`. Provide the receiver's IP address and the full path to the file on the **sending computer**:

```powershell
.\p2p-share.exe `
  -listen :4000 `
  -peer 192.168.1.25:3000 `
  -send "C:\Users\Alice\Desktop\report.pdf"
```

Replace `192.168.1.25` with the receiver's actual IPv4 address and replace the example file path with the file you want to send. Keep the quotes around paths that contain spaces. The sending node waits up to 30 seconds for a peer connection, sends the file, waits for the receiver to confirm it was written, then exits.

The receiver writes the file to:

```text
received_files\<original-file-name>
```

This folder is created under the receiver process's current working directory (normally the folder from which `p2p-share.exe` was launched). If a file with the same name already exists there, it is overwritten.

## Command-line options

| Option | Default | Description |
| --- | --- | --- |
| `-listen` | `:3000` | TCP address for this node to listen on. Use a port such as `:4000`. |
| `-peer` | empty | Optional address of the peer to connect to, in `IP:port` form, such as `192.168.1.25:3000`. |
| `-send` | empty | Optional path to a file to send after a peer connects. When omitted, the node stays running as a receiver/listener. |

The `-peer` setting is used by the sending node to dial the receiver. The receiver does not need a `-peer` option for a simple two-computer transfer.

## Check connectivity

From the sending computer, test whether the receiver's port is reachable:

```powershell
Test-NetConnection 192.168.1.25 -Port 3000
```

Look for:

```text
TcpTestSucceeded : True
```

If it is `False`, verify that both computers are on a network that permits device-to-device connections, the receiver is running, the IP and port are correct, and Windows Firewall allows inbound TCP on that port.

## Run a local loopback test

You can test the transfer on one computer using two PowerShell windows. In the first:

```powershell
.\p2p-share.exe -listen :3000
```

In the second, replace the file path with a real local file:

```powershell
.\p2p-share.exe -listen :4000 -peer 127.0.0.1:3000 -send "C:\path\to\test.txt"
```

The file should appear in `received_files` under the first window's working directory.

## How it works

1. The receiver opens a TCP listener on its configured port.
2. The sender opens its own listener and dials the receiver using `-peer`.
3. The sender sends file metadata followed by the file bytes over the TCP peer connection.
4. The receiver saves the bytes using the original base file name in `received_files`.
5. The receiver sends a confirmation message; after receiving it, the sender exits.

## Troubleshooting

### `.\p2p-share.exe` is not recognized

PowerShell looks for `.\p2p-share.exe` in the current directory. Change to the directory containing the executable first, or run it using its full path:

```powershell
& "C:\path\to\p2p-share.exe" -listen :3000
```

### `open ... The system cannot find the path specified`

The value passed to `-send` must be the actual path to an existing file on the sending computer. In File Explorer, use **Copy as path** and paste it after `-send` (keeping quotes if the path contains spaces).

### Timed out waiting for a peer connection

The sender did not establish a connection within 30 seconds. Check the receiver's IP and port, confirm that the receiver is still running, and test the port with `Test-NetConnection`. Check firewall rules and make sure the Wi-Fi network does not isolate wireless clients.

### The receiver does not get the file

Keep the receiver running until the sender reports that the file was sent. Look in the receiver's `received_files` folder relative to the directory from which it was started. Confirm that the receiver has permission to create files there and check both terminal windows for errors.

### Testing across different networks

Direct TCP connections across the internet commonly fail because of NAT and router firewalls. This project does not provide NAT traversal, a relay, or automatic discovery. Use a trusted VPN that places both computers on a reachable private network for a test; do not expose the port directly to the public internet.

## Current limitations

- One file is sent per sender invocation; there is no interactive file browser, resume support, progress bar, or multi-file batch mode.
- The sender and receiver must be able to connect directly over TCP. There is no relay service or NAT traversal.
- Peer connections are unauthenticated and network data is not protected by TLS. Anyone able to connect to the listening port may interact with the protocol.
- Received files use the sender-provided base name. A same-named file in the receiving folder is overwritten.
- The application is a prototype and should not be used as a secure general-purpose file-sharing service.
