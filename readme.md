# Nomsys

A terminal system resource monitor built with Go. Displays live CPU, memory, network, and uptime stats alongside the top 20 processes by CPU and memory usage. Inspired by htop

## Running

```bash
go run ./cmd/cli/main.go
```

### Flags

| Flag | Default | Description                                    |
| ---- | ------- | ---------------------------------------------- |
| `-d` | `2`     | Delay between updates in seconds               |
| `-n` | `0`     | Exit after N iterations (0 = run indefinitely) |

## Keybindings

| Key      | Action                              |
| -------- | ----------------------------------- |
| `Tab`    | Switch focus between process tables |
| `↑ / ↓`  | Scroll through process list         |
| `q`      | Quit                                |
| `Ctrl+C` | Quit                                |

## Stats

- **Memory** — used, free, total, % used
- **CPU** — per-core usage %
- **Network** — total bytes sent/received, active TCP connections
- **Uptime** — days, hours, minutes
- **Top 20 by CPU** — name, PID, user, CPU %, start time, status
- **Top 20 by Memory** — name, PID, user, memory %, start time, status
