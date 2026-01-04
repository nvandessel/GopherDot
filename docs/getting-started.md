# Getting Started with go4dot

This guide will help you get up and running with go4dot, whether you're setting up a new machine or creating a new dotfiles repository.

## 🏁 Setting Up a New Machine

If you already have a dotfiles repository configured with go4dot:

1. **Clone your repository:**
   ```bash
   git clone https://github.com/yourusername/dotfiles.git ~/dotfiles
   cd ~/dotfiles
   ```

2. **Install go4dot:**
   (If you haven't already, see [Installation](installation.md))
   ```bash
   curl -fsSL https://raw.githubusercontent.com/nvandessel/go4dot/main/scripts/install.sh | bash
   ```

3. **Run the install command:**
   ```bash
   g4d install
   ```

   The interactive wizard will guide you through:
   - Installing system dependencies (git, zsh, neovim, etc.)
   - Configuring machine-specific settings (git user/email, GPG keys)
   - Selecting which configs to stow (e.g. core vs optional)
   - Cloning external dependencies (plugins, themes)

## 🆕 Creating New Dotfiles

If you have existing dotfiles but haven't used go4dot before:

1. **Navigate to your dotfiles directory:**
   ```bash
   cd ~/path/to/your/dotfiles
   ```

2. **Initialize configuration:**
   ```bash
   g4d init
   ```

   go4dot will scan your directory for common config folders (nvim, git, zsh, tmux, etc.) and generate a `.go4dot.yaml` file.

3. **Customize your config:**
   Edit `.go4dot.yaml` to fine-tune your setup. See [Configuration Reference](config-reference.md) for details.

4. **Test your setup:**
   ```bash
   g4d install
   ```

## 🏥 Health Checks

Run `g4d doctor` at any time to verify your installation. It checks for:
- Missing system dependencies
- Broken symlinks
- Misconfigured external dependencies
- Valid machine configuration

```bash
g4d doctor
```

## 🔄 Updating

To pull the latest changes from your dotfiles repo and apply them:

```bash
g4d update
```

This will:
1. `git pull` in your dotfiles directory
2. Detect new configs or changes
3. Re-run `stow` to ensure symlinks are correct
4. Update external dependencies (plugins, themes)

## 🧹 Maintenance

- **List installed configs:**
  ```bash
  g4d list
  ```

- **Reconfigure machine settings:**
  (Useful if you want to change your git email or GPG key)
  ```bash
  g4d reconfigure
  ```

- **Uninstall:**
  (Removes symlinks, keeps the files)
  ```bash
  g4d uninstall
  ```
