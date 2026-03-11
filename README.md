# irlearn

> Capture once. Replay forever.

Point your IR blaster at any remote, press a button, and `irlearn` captures the code instantly.
Organise commands into named sessions, replay them on demand, and script headless capture into your automation pipelines — all from the terminal.

```
┌─────────────────────────────┬──────────────────────┐
│  Buffer                     │  Preview             │
│  > msg 1  (42 bytes)        │  Name: —             │
│    msg 2  (38 bytes)        │  Bytes: 42           │
│    msg 3  (42 bytes)        │  Hex: 4A 2B 00 FF …  │
├─────────────────────────────┤                      │
│  Session: my-tv.yaml        │                      │
│  > power                    │                      │
│    volume_up                │                      │
│    mute                     │                      │
└─────────────────────────────┴──────────────────────┘
│ MQTT: connected  [r]eplay [s]ave [d]elete [c]lear  │
│ :                                                   │
└─────────────────────────────────────────────────────┘
```

## Quick Start

```bash
# Build
go install github.com/dadao/zigbee2mqtt-ir-blaster-learner@latest

# Create config (~/.config/mqttirlearn/config.yaml)
mkdir -p ~/.config/mqttirlearn
cat > ~/.config/mqttirlearn/config.yaml << EOF
mqtt:
  host: 192.168.1.100
  port: 1883
  user: ""
  password: ""
device: MY_IR_BLASTER
EOF

# Launch the TUI
irlearn ui
```

All config values can also be passed as flags — run `irlearn --help` for the full list.

## Features

- **Live capture** — receives IR codes the moment your device learns them
- **Session management** — group commands into named YAML files (e.g. `my-tv.yaml`)
- **Instant replay** — test any captured code directly from the TUI
- **Lock mode** — hands-free loop: auto-resends the learn command after each capture
- **Headless capture** — `irlearn capture` prints the base64 code to stdout for scripting
- **Flag overrides** — any config value can be overridden per-invocation without editing files

## Documentation

| Document | Audience |
|----------|----------|
| [Usage Manual](docs/USAGE_MANUAL.md) | End users — setup, TUI guide, keyboard shortcuts, troubleshooting |
| [Architecture](docs/ARCHITECTURE.md) | Developers — package map, data flow, dependency injection |
| [Feature Reference](docs/FEATURES.md) | Quick reference — MQTT payloads, session schema |

## License

MIT
