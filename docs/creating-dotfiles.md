# Creating Your Own Dotfiles

This guide explains how to structure your dotfiles repository for use with go4dot.

## Directory Structure

go4dot relies on [GNU Stow](https://www.gnu.org/software/stow/), which manages symlinks based on directory structure.

Each configuration group should be in its own directory. Inside that directory, mimic the structure of your home folder.

### Example

Suppose you want to manage `~/.zshrc` and `~/.config/nvim/init.lua`.

**Directory Structure:**

```
my-dotfiles/
├── .go4dot.yaml        # The manifest
├── zsh/                # Config Group: zsh
│   └── .zshrc          # -> symlinks to ~/.zshrc
└── nvim/               # Config Group: nvim
    └── .config/
        └── nvim/
            └── init.lua # -> symlinks to ~/.config/nvim/init.lua
```

When go4dot stows `zsh`, it symlinks `my-dotfiles/zsh/.zshrc` to `~/.zshrc`.
When go4dot stows `nvim`, it symlinks `my-dotfiles/nvim/.config/nvim/init.lua` to `~/.config/nvim/init.lua`.

## Step-by-Step Guide

1. **Create a directory** for your dotfiles:
   ```bash
   mkdir ~/my-dotfiles
   cd ~/my-dotfiles
   git init
   ```

2. **Move existing config files** into the repo, maintaining structure:
   ```bash
   # Move .zshrc
   mkdir zsh
   mv ~/.zshrc zsh/
   
   # Move nvim config
   mkdir -p nvim/.config/nvim
   mv ~/.config/nvim/init.lua nvim/.config/nvim/
   ```

3. **Initialize go4dot config:**
   ```bash
   g4d init
   ```
   Follow the prompts to generate `.go4dot.yaml`.

4. **Install/Link them back:**
   ```bash
   g4d install
   ```
   This will create the symlinks pointing back to your files in `~/my-dotfiles`.

5. **Commit and Push:**
   ```bash
   git add .
   git commit -m "Initial commit"
   # Add your remote and push...
   ```

## Best Practices

- **Keep it modular**: Separate configs by tool (git, zsh, vim, tmux). This allows you to install only what you need on different machines.
- **Use Machine Configs**: Don't commit secrets or machine-specific paths. Use the `machine_config` feature in `.go4dot.yaml` to generate a local file (e.g. `~/.gitconfig.local`) that you include in your main config.
- **External Deps**: Use the `external` section to manage plugins (like TPM, zsh plugins) instead of using git submodules, which can be finicky.
