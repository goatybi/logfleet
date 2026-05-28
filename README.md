# logfleet

view logs from all your servers in a single terminal. No agents, no databases, no setup.

```
[prod-web    ] [access.log          ] 192.168.1.1 GET /api/users 200
[prod-db     ] [postgresql.log      ] connection received: host=10.0.0.1
[prod-web    ] [error.log           ] upstream timed out
[staging     ] [syslog              ] deploy finished v2.3.1
```

## Install

### macOS (Apple Silicon — M1/M2/M3/M4)
```bash
curl -L https://github.com/goatybi/logfleet/releases/latest/download/logfleet-macos-arm64 -o logfleet
chmod +x logfleet
xattr -dr com.apple.quarantine logfleet
sudo mv logfleet /usr/local/bin/
```

### macOS (Intel)
```bash
curl -L https://github.com/goatybi/logfleet/releases/latest/download/logfleet-macos-intel -o logfleet
chmod +x logfleet
xattr -dr com.apple.quarantine logfleet
sudo mv logfleet /usr/local/bin/
```

### Linux (amd64)
```bash
curl -L https://github.com/goatybi/logfleet/releases/latest/download/logfleet-linux-amd64 -o logfleet
chmod +x logfleet
sudo mv logfleet /usr/local/bin/
```

### Linux (arm64)
```bash
curl -L https://github.com/goatybi/logfleet/releases/latest/download/logfleet-linux-arm64 -o logfleet
chmod +x logfleet
sudo mv logfleet /usr/local/bin/
```

### Windows
Install [WSL](https://learn.microsoft.com/en-us/windows/wsl/install) first, then use Linux instructions above inside WSL terminal.

### build from source (Go 1.22+ required)
```bash
git clone https://github.com/goatybi/logfleet
cd logfleet
go build -o logfleet ./cmd/logfleet/
```

## quick start

**1. create config:**
```bash
logfleet init
```

**2. add your servers:**
```bash
nano ~/.logfleet/config.yaml
```

fill in your server details:
```yaml
servers:
  - name: prod-web
    host: 1.2.3.4
    user: root
    key: ~/.ssh/id_rsa
    logs:
      - /var/log/nginx/access.log
      - /var/log/nginx/error.log
      - /var/log/syslog

  - name: prod-db
    host: 5.6.7.8
    user: ubuntu
    key: ~/.ssh/id_rsa
    logs:
      - /var/log/postgresql/postgresql-16-main.log
```

Save and exit: **Ctrl+O** → Enter → **Ctrl+X**

**3. check servers are alive:**
```bash
logfleet status
```

**4. stream logs:**
```bash
logfleet
```

## Commands

```bash
logfleet                          # stream all logs
logfleet tail --level error       # errors only (fatal, panic, critical...)
logfleet tail --level warn        # warnings and errors
logfleet tail --server prod-web   # one server only
logfleet tail --grep "timeout"    # filter by text
logfleet status                   # check all servers are alive
logfleet status --timeout 5s      # with custom timeout
logfleet init                     # create example config
```

combine filters:
```bash
logfleet tail --level error --server prod-web
logfleet tail --level warn --grep "database"
```

stop with **Ctrl+C**.

## how it works

- connects to each server via SSH (your key, no passwords)
- runs `tail -F` on every log file
- merges all streams into one terminal with color prefixes
- auto-reconnects if connection drops (after 5 sec)
- lines with `error/fatal/panic` highlighted red, `warn/timeout` yellow

**no agents on servers. Just SSH.**

## requirements

- SSH key without passphrase (or ssh-agent)
- `tail` on remote servers (available everywhere)

## roadmap

- [x] `--level error` — errors only
- [x] `logfleet status` — check all servers alive
- [ ] `--since 1h` — show logs from specific time
- [ ] known_hosts support
- [ ] save logs to file
