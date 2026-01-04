# Advanced Dotfiles Example

This example demonstrates a more complex setup including:
- Multiple tools (git, tmux, nvim)
- External dependencies (TPM, LazyVim)
- Complex machine configuration (conditional GPG signing)
- Optional configurations
- Platform-specific packages

## Structure

```
.
├── .go4dot.yaml        # Configuration manifest
├── git/
│   └── .gitconfig      # Symlinks to ~/.gitconfig
├── nvim/
│   └── init.lua        # Symlinks to ~/.config/nvim/init.lua (if path adjusted)
└── tmux/
    └── .tmux.conf      # Symlinks to ~/.tmux.conf
```

## Highlights

- **External Deps**: Shows how to clone TPM (Tmux Plugin Manager).
- **Machine Config**: Shows how to use Go templates to conditionally add GPG signing config.
- **Dependencies**: Shows how to map package names (e.g. `fd` vs `fd-find`).
