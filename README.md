# 💎 Better Paping

Hey! This is **Better Paping**, a single tool that combines the best parts of TCP pinging, Web pinging, and Minecraft server pinging. No more swapping between different scripts,everything is right here.

---

## 🚀 Quick Start

Build it once, use it everywhere:

```bash
go build -o betterpaping.exe betterpaping.go
betterpaping.exe <mode> [options]
```

### 📡 TCP Mode (Classic)
Just like the original but with better stats and color-coded latency.
```bash
./better_paping.exe tcp 1.1.1.1 80 500 0
```

### 🌐 Web Mode (HTTP/HTTPS)
Track if your site is actually up and how fast it's responding to real headers.
```bash
./better_paping.exe web https://google.com 1000 60
```

### 🎮 MC Mode (Minecraft)
The best diagnostic for MC servers. Auto-resolves SRVs (so you don't need the port) and checks the protocol version.
```bash
./better_paping.exe mc play.hypixel.net 25565 650 0
```

---

## ✨ Features that actually matter

- **Color Logic:** Red for slow, yellow for "meh", green for fast. You can see network issues at a glance.
- **ASCII Art & Stats:** Clean, icon-based summaries so you don't have to squint at raw text.
- **SRV Support:** MCPing handles those annoying hidden Minecraft ports for you.
- **Human Coded:** No enterprise bloat. Just direct, fast, and messy enough to be real.

---

## 🛠️ Build & Install

Make sure you have Go installed, then hit this:
```bash
go get github.com/fatih/color
go build -o better_paping.exe paping_suite.go
```

---

*Made for speed and clarity. No BS.*
