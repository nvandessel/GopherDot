# Installation Guide

go4dot can be installed in several ways. Choose the method that works best for you.

## ⚡ Quick Install (Recommended)

The easiest way to install go4dot is using the automated install script. This will download the latest release for your platform and install it to `/usr/local/bin` (or `~/.local/bin` if you don't have sudo access).

```bash
curl -fsSL https://raw.githubusercontent.com/nvandessel/go4dot/main/scripts/install.sh | bash
```

To verify the installation:

```bash
g4d version
```

## 📦 Binary Releases

You can manually download the latest binary for your platform from the [GitHub Releases](https://github.com/nvandessel/go4dot/releases) page.

1. Download the archive for your OS and architecture (e.g., `g4d-linux-amd64.tar.gz`).
2. Extract the archive:
   ```bash
   tar -xzf g4d-linux-amd64.tar.gz
   ```
3. Move the binary to a directory in your PATH:
   ```bash
   sudo mv g4d /usr/local/bin/
   ```
4. Make it executable:
   ```bash
   sudo chmod +x /usr/local/bin/g4d
   ```

## 🐹 Install with Go

If you have Go installed (version 1.21+), you can install go4dot directly from the source:

```bash
go install github.com/nvandessel/go4dot/cmd/g4d@latest
```

Ensure your `$GOPATH/bin` is in your `$PATH`.

## 🏗️ Build from Source

To build from source, you'll need Go 1.21+ and `make`.

1. Clone the repository:
   ```bash
   git clone https://github.com/nvandessel/go4dot.git
   cd go4dot
   ```

2. Build the binary:
   ```bash
   make build
   ```

3. Install locally:
   ```bash
   make install
   ```

## 🗑️ Uninstallation

To remove go4dot:

```bash
sudo rm /usr/local/bin/g4d
# Or if installed in ~/.local/bin
rm ~/.local/bin/g4d
```

To remove the state file and logs:

```bash
rm -rf ~/.config/go4dot
```
