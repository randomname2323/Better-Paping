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
Just like the original but with better stats and color coded latency.
```bash
./betterpaping.exe tcp 1.1.1.1 80 500 0
```

### 🌐 Web Mode (HTTP/HTTPS)
Track if your site is actually up and how fast it's responding.
```bash
./betterpaping.exe web https://google.com 1000 60
```

### 🎮 MC Mode (Minecraft)
The best paping tool for Minecraft servers. Auto resolves SRVs (so you don't need the port) and checks the protocol version.
```bash
./betterpaping.exe mc play.hypixel.net 25565 650 0
```

---

## ✨ Features that actually matter

- **Color Logic:** Red for slow, yellow for "meh", green for fast. You can see network issues very cleanly.
- **ASCII Art & Stats:** Clean, so you don't have to squint at raw text.
- **SRV Support:** Minecraft Ping handles those annoying hidden Minecraft ports for you.

---

## 🛠️ Build & Install

Make sure you have Go installed, then hit this:
```bash
go get github.com/fatih/color
go build -o betterpaping.exe betterpaping.go
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
*Built with ❤️ by Zyre*
