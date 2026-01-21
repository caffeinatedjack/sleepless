# Sleepless

A collection of personal productivity tools for managing your digital life through your Vim wiki.

## Executables

### regimen

Personal knowledge management and productivity CLI for your Vim wiki.

**Features:**
- 📝 **Notes**: Daily journaling and floating notes with full-text search
- 🎯 **Goals**: Task and goal tracking with hierarchical organization
- 🍳 **Recipes**: Recipe collection and management
- 🔐 **Encryption**: Wiki-wide encryption with AES-256-GCM
- 📋 **Templates**: Structured note templates for meetings, reflections, and more
- 🌍 **World Clock**: Time zone conversion and world clock

**Installation:**
```bash
go build -o regimen ./cmd/regimen
```

**Basic Usage:**
```bash
# Set your wiki directory (optional, defaults to ~/wiki)
export REGIMEN_WIKI_DIR=~/wiki

# Notes
regimen note add "Had an interesting idea today"
regimen note add --template meeting
regimen note today
regimen note search "idea"

# Goals
regimen goals add "Complete project documentation"
regimen goals list
regimen goals done <id>

# Recipes
regimen recipes add pasta
regimen recipes list

# Encryption (optional)
regimen encrypt
regimen decrypt

# World clock
regimen when
```

## Commands

### `regimen note` - Note Management

Manage daily notes and floating notes in your wiki.

**Subcommands:**
- `add [text]` - Add a note (opens editor if no text provided)
- `today` - Show today's daily note
- `week` - Show notes from the past 7 days
- `month` - Show notes from current month
- `list` - List recent notes
- `show <id|date>` - Display a specific note
- `search <query>` - Search notes by text
- `edit <id|date>` - Edit an existing note
- `delete <id|date>` - Delete a note
- `tags` - List all tags or show notes with specific tag
- `tag <id> <tags>` - Add tags to a note
- `untag <id> <tags>` - Remove tags from a note
- `random` - Display random note(s)
- `report` - Generate work report from recent notes
- `stats` - Show statistics about your notes
- `template` - Manage note templates

**Examples:**
```bash
# Quick daily note
regimen note add "Met with team to discuss Q1 roadmap" --tags work,meeting

# Floating note for research
regimen note add --floating "Research on distributed caching" --tags research,tech

# Create note from template
regimen note add --template meeting

# Search and filter
regimen note search "roadmap"
regimen note tags work
regimen note list --limit 5

# View and analyze
regimen note today
regimen note week
regimen note stats
regimen note random --count 3
```

**Note Types:**
- **Daily notes**: Date-based notes (YYYY-MM-DD.md) with timestamped sections
- **Floating notes**: Standalone notes with unique 8-character hex IDs

**Templates:**
Built-in templates available:
- `meeting` - Meeting notes with agenda and action items
- `reflection` - Daily reflection prompts
- `idea` - Idea capture template
- `report` - Status report template

### `regimen goals` - Goal Management

Track tasks and goals with hierarchical organization.

**Subcommands:**
- `add <title>` - Create a new goal
- `list` - List all goals
- `done <id>` - Mark a goal as complete
- `view tree` - View goals in tree structure

**Examples:**
```bash
regimen goals add "Write documentation" --topic work --priority high
regimen goals list
regimen goals done abc123
```

### `regimen recipes` - Recipe Management

Manage cooking recipes in your wiki.

**Subcommands:**
- `add <name>` - Add a new recipe
- `list` - List all recipes
- `show <name>` - Display a recipe

**Examples:**
```bash
regimen recipes add pasta
regimen recipes list
regimen recipes show pasta
```

### `regimen encrypt` / `regimen decrypt` - Wiki Encryption

Encrypt your entire wiki for security.

**Encryption:**
```bash
# Encrypt wiki (interactive password prompt)
regimen encrypt

# Or use stdin for scripting
echo "my-password" | regimen encrypt --passphrase-stdin
```

**Decryption:**
```bash
# Decrypt wiki (interactive password prompt)
regimen decrypt

# Or use stdin for scripting
echo "my-password" | regimen decrypt --passphrase-stdin
```

**Features:**
- AES-256-GCM encryption
- Argon2id key derivation (time=3, memory=128MiB, threads=4)
- Encrypts `.md` and `.json` files
- Automatically skips `.git/` directory
- Per-file authentication prevents tampering
- All commands blocked when wiki is encrypted

### `regimen when` - World Clock

Display world clock and convert times between zones.

```bash
regimen when
```

### `regimen decide` - Random Choice

Random choice utilities for decision making.

```bash
regimen decide flip        # Coin flip
regimen decide roll 6      # Roll a die
```

### `regimen banner` - ASCII Art

Display themed ASCII art banners.

```bash
regimen banner
```

## Global Flags

- `--wiki-dir <path>` - Specify wiki directory (default: `~/wiki`, env: `REGIMEN_WIKI_DIR`)
- `-h, --help` - Show help
- `-v, --version` - Show version

## Configuration

### Environment Variables

- `REGIMEN_WIKI_DIR` - Default wiki directory location
- `EDITOR` - Preferred text editor for editing notes (default: `vim`)

### Wiki Structure

```
~/wiki/
├── notes/                  # Note storage
│   ├── 2026-01-21.md      # Daily notes (YYYY-MM-DD)
│   └── a1b2c3d4.md        # Floating notes (hex ID)
├── .templates/             # Note templates
│   ├── meeting.md
│   ├── reflection.md
│   ├── idea.md
│   └── report.md
├── recipes/                # Recipe files
│   └── *.md
├── tasks/                  # Task/goal data
│   └── tasks.json
└── .encrypted              # Encryption marker (when encrypted)
```

## Note Format

Notes use Markdown with YAML frontmatter:

```markdown
---
id: a1b2c3d4
type: floating
tags:
  - work
  - idea
created: 2026-01-21T19:00:00Z
modified: 2026-01-21T19:00:00Z
---

# My Note Title

Note content goes here.

You can use #inline-tags too.
```

## Templates

Templates support placeholders:
- `{{DATE}}` - Current date (YYYY-MM-DD)
- `{{TIME}}` - Current time (HH:MM)
- `{{DATETIME}}` - Current date and time
- `{{PROMPT:question}}` - Interactive prompt for user input

Example template:
```markdown
# Meeting: {{PROMPT:Meeting title}}

**Date:** {{DATE}}
**Time:** {{TIME}}

## Notes


## Action Items
- [ ] 
```

## Development

**Build:**
```bash
go build -o regimen ./cmd/regimen
```

**Test:**
```bash
go test ./...
```

**Run tests with coverage:**
```bash
go test ./... -cover
```

## License

See LICENSE file for details.

## Contributing

See CONTRIBUTING.md for development guidelines and project rules.
