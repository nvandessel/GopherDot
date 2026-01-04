# Command Reference

## `g4d install`
The main entry point. Orchestrates the full setup process.
- **Usage**: `g4d install [path]`
- **Flags**:
  - `--auto`: Run in non-interactive mode using defaults.
  - `--minimal`: Install only core configs/deps, skip optional ones.
  - `--skip-deps`: Skip system dependency check/install.
  - `--skip-external`: Skip cloning external dependencies.
  - `--skip-machine`: Skip machine configuration prompts.
  - `--skip-stow`: Skip stowing dotfiles.

## `g4d init`
Bootstrap a new configuration from existing dotfiles.
- **Usage**: `g4d init [path]`
- **Description**: Scans the directory for config folders and interacts with you to generate a `.go4dot.yaml`.

## `g4d doctor`
Check the health of your installation.
- **Usage**: `g4d doctor`
- **Checks**:
  - System dependencies
  - Broken symlinks
  - Missing external dependencies
  - Machine config validity

## `g4d update`
Update dotfiles and external dependencies.
- **Usage**: `g4d update`
- **Actions**:
  - `git pull` in dotfiles repo
  - Restow configs
  - Update external git repos

## `g4d list`
List all available and installed configurations.
- **Usage**: `g4d list`
- **Flags**:
  - `-a, --all`: Show all details including archived/hidden.

## `g4d reconfigure`
Re-run machine-specific configuration prompts.
- **Usage**: `g4d reconfigure [id]`
- **Description**: Useful if you need to change a value (like git email) without reinstalling everything.

## `g4d uninstall`
Remove symlinks and clean up.
- **Usage**: `g4d uninstall`
- **Flags**:
  - `-f, --force`: Skip confirmation.
- **Description**: Unstows all configs. Does **not** delete your actual dotfiles files, only the symlinks.

## `g4d detect`
Show platform information.
- **Usage**: `g4d detect`
- **Output**: OS, Distro, Package Manager, etc.

## `g4d stow`
Manual stow operations.
- `g4d stow add <config>`: Stow a specific config group.
- `g4d stow remove <config>`: Unstow a specific config group.
- `g4d stow refresh`: Restow all active configs.

## `g4d external`
Manage external dependencies manually.
- `g4d external status`: Show status of external repos.
- `g4d external clone [id]`: Clone specific repo.
- `g4d external update [id]`: Update specific repo.
- `g4d external remove <id>`: Remove specific repo.

## `g4d machine`
Manage machine configuration manually.
- `g4d machine status`: Show status of machine configs.
- `g4d machine show <id>`: Preview generated config.
- `g4d machine configure [id]`: Run prompts for specific config.
