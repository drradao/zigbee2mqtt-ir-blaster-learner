# Usage Manual

End-user guide for setting up and operating `irlearn`.

---

## Prerequisites

- **zigbee2mqtt** running with an IR blaster device paired (e.g. Moes UFO-R11)
- **MQTT broker** (e.g. Mosquitto) accessible from your machine
- **Go** (for building from source)

---

## Installation

### From source

```bash
go install github.com/dadao/zigbee2mqtt-ir-blaster-learner@latest
```

The binary is installed to `$(go env GOPATH)/bin/irlearn`. Make sure that directory is on
your `PATH`.

### Build locally

```bash
git clone https://github.com/dadao/zigbee2mqtt-ir-blaster-learner
cd zigbee2mqtt-ir-blaster-learner
go build -o irlearn .
```

---

## Configuration

### Config file location

```
~/.config/mqttirlearn/config.yaml
```

Create it before first run:

```bash
mkdir -p ~/.config/mqttirlearn
```

### Full config example

```yaml
mqtt:
  host: 192.168.1.100   # MQTT broker hostname or IP
  port: 1883            # MQTT broker port (1883 = plain, 8883 = TLS)
  user: ""              # MQTT username (leave empty if broker has no auth)
  password: ""          # MQTT password

device: living_room_ir  # zigbee2mqtt friendly name of your IR blaster
```

All fields can be overridden per-invocation with CLI flags (see below). This is useful for
switching between devices or brokers without editing the file.

### CLI flag overrides

Every config field has a corresponding flag. Only explicitly-provided flags override the
file; unset flags leave the file value intact.

| Flag                   | Description                          |
|------------------------|--------------------------------------|
| `--config PATH`        | Use a different config file          |
| `--mqtt-host HOST`     | MQTT broker host                     |
| `--mqtt-port PORT`     | MQTT broker port                     |
| `--mqtt-user USER`     | MQTT username                        |
| `--mqtt-password PASS` | MQTT password                        |
| `--device NAME`        | zigbee2mqtt device friendly name     |
| `--session FILE`       | Session file path                    |
| `--log FILE`           | Write logs to file instead of stderr |

---

## First Run

If the config file is absent or any required field (`host`, `device`) is missing, `irlearn`
automatically shows an interactive setup form before attempting to connect.

Fill in the required fields and press `Tab` or `Enter` to move between them. When all fields
are complete you are asked `Save to config? [y/N]`. Press `y` to write the values to the
config file; press `n` or `Enter` to proceed without saving (useful when you prefer to set
values via CLI flags each time).

Press `Esc` at any point to abort.

---

## The `config` Command

Opens the full configuration form, even when all fields are already set. Use this to review
or change any setting interactively.

```bash
irlearn config
irlearn config --config /path/to/other.yaml   # edit a specific config file
```

Fields are pre-populated with the current config values. Navigate with `Tab`/`Shift+Tab` and
confirm with `Enter`. You are asked whether to save on exit.

---

## The `ui` Command

The interactive TUI is the primary interface.

```bash
irlearn ui
irlearn ui --session my-tv.yaml      # load a specific session file
irlearn ui --device bedroom_ir       # override device for this session
irlearn ui --login                   # prompt for MQTT credentials before connecting
```

### Layout

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
│ :                                                  │
└────────────────────────────────────────────────────┘
```

| Pane                      | What it shows                                                       |
|---------------------------|---------------------------------------------------------------------|
| **Buffer** (top-left)     | Live stream of IR codes received from the device, most-recent first |
| **Session** (bottom-left) | Commands saved to the current session file                          |
| **Preview** (right)       | Decoded byte length and hex preview of the selected item            |
| **Status bar** (bottom)   | MQTT connection state and keyboard hint strip                       |
| **Prompt line**           | Vim-style `:` prompt, appears when naming a command                 |

### Keyboard Shortcuts

| Key            | Action                                                  |
|----------------|---------------------------------------------------------|
| `↑` / `k`      | Move selection up                                       |
| `↓` / `j`      | Move selection down                                     |
| `Tab`          | Switch focus between Buffer and Session panes           |
| `r`            | Replay selected item (test the IR code live)            |
| `s`            | Save selected buffer item to session (prompts for name) |
| `n`            | Rename selected session command                         |
| `d`            | Delete selected item                                    |
| `C`            | Clear all items from the buffer                         |
| `l`            | Send learn-mode command to device                       |
| `L`            | Toggle lock mode (see below)                            |
| `q` / `Ctrl+C` | Quit                                                    |

### Step-by-step: Learning a button

1. Run `irlearn ui`
2. Press `l` to put the IR blaster into learn mode
3. Point your remote at the IR blaster and press the button you want to capture
4. The IR code appears in the **Buffer** pane
5. With the code selected, press `s`
6. Type a name (e.g. `power`) at the `:` prompt and press `Enter`
7. The command appears in the **Session** pane and is written to the session file immediately

### Step-by-step: Replaying a command

1. Navigate to the **Session** pane with `Tab`
2. Select the command with `↑`/`↓`
3. Press `r` — the IR code is sent to the device

### Lock mode

Press `L` to enable lock mode. In lock mode, `irlearn` automatically re-sends the learn
command to the device every time it receives a new IR code. This creates a continuous capture
loop — useful when learning many buttons in sequence without pressing `l` between each one.

Press `L` again to disable.

---

## The `capture` Command

For scripting and automation — captures one IR code and prints the base64 payload to stdout.

```bash
irlearn capture [flags]
```

### Flags

| Flag          | Default | Description                                                   |
|---------------|---------|---------------------------------------------------------------|
| `--timeout N` | `30`    | Seconds to wait for an IR code before giving up               |
| `--verbose`   | false   | Print progress messages to stderr                             |
| `--login`     | false   | Prompt for MQTT credentials via stdin before connecting       |

All global flags (`--device`, `--mqtt-host`, etc.) also apply.

> **Note on `--login`:** `capture --login` prompts for username and password on the terminal
> before connecting. The input is not echo-suppressed (the password is visible). If you need
> hidden password entry, set `mqtt.password` in your config file or use the `--mqtt-password`
> flag instead.

### Exit codes

| Code | Meaning                                              |
|------|------------------------------------------------------|
| `0`  | IR code received and printed to stdout               |
| `1`  | Timeout — no code received within the timeout window |

### Example — save a code to a file

```bash
irlearn capture --device my_ir --verbose > captured.b64
```

### Example — trigger learn mode first, then capture

```bash
# In shell: publish learn command manually, then wait for code
irlearn capture --timeout 60
```

---

## Session Files

### Format

```yaml
device: living_room_tv
commands:
  - name: power
    payload: <base64-encoded IR code>
    captured_at: 2024-01-01T00:00:00Z
  - name: volume_up
    payload: <base64-encoded IR code>
    captured_at: 2024-01-01T00:00:01Z
```

### Default location

`session.yaml` in the current working directory.

Override with `--session path/to/file.yaml` or set a path in the config file.

### When files are written

Session files are written immediately when you save or rename a command, or delete a session
command. There is no separate "save" step — the file is always up to date.

---

## Troubleshooting

**Cannot connect to MQTT broker**
- Verify `host` and `port` in your config
- Check that the broker is reachable: `nc -zv 192.168.1.100 1883`
- If the broker requires authentication, set `user` and `password`

**No IR code appears in the buffer**
- Confirm the device is in learn mode (press `l` or check zigbee2mqtt logs)
- Verify `device` matches the **friendly name** in zigbee2mqtt exactly (case-sensitive)
- Check zigbee2mqtt logs to confirm the device is publishing on `zigbee2mqtt/DEVICE`

**Session file not saving**
- Check that the directory containing the session file exists and is writable
- Use `--log /tmp/irlearn.log` and inspect the log for write errors
