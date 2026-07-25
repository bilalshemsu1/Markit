# Markit

![Version](https://img.shields.io/badge/version-v1.7.0-blue) ![Platform](https://img.shields.io/badge/platform-Windows-lightgrey) ![License](https://img.shields.io/badge/license-MIT-green)

A **fast**, distraction-free Markdown editor for Windows built with Wails. Features a clean interface, split-view editing, native file dialogs, and keyboard shortcuts. No server. No account. Your files never leave your device.

## ✨ Features

- **Live Preview** — Edit and preview your Markdown side-by-side in real-time
- **Split View** — Choose between Edit-only, Split, or Preview-only modes
- **Document Outline** — Live outline panel showing all headings with click-to-navigate
- **File Watching** — Automatic detection of external file changes with conflict resolution
- **Zoom Controls** — Adjust the editor zoom level (50% - 250%) with buttons or keyboard shortcuts
- **Collapsible Sidebar** — Manage multiple files with a collapsible file sidebar
- **Native File Dialogs** — Open and save files directly to disk using Windows native dialogs
- **Syntax Highlighting** — Code blocks with syntax highlighting
- **Keyboard Shortcuts** — Full keyboard support for common actions
- **Drag & Drop** — Drag files directly into the editor to open them
- **Startup File Support** — Open a file directly from command line
- **Clean UI** — Professional icon-based interface with smooth animations

## 📦 Installation

### Pre-built Binaries

Download the latest release from the [Releases page](https://github.com/bilalshemsu1/md-file-reader/releases).

### Build from Source

**Prerequisites:**
- Go 1.21 or later
- Node.js 18 or later
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

**Build Steps:**
```bash
# Clone the repository
git clone https://github.com/bilalshemsu1/md-file-reader.git
cd md-file-reader

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build the application
wails build
```

## 🚀 Usage

### Quick Start

1. Launch Markit
2. Click **Open** or drag `.md` files onto the window
3. Edit using the toolbar or keyboard shortcuts
4. Switch between Edit / Split / Preview modes (top-right)
5. Click **Save** to save directly to disk

### Command Line Usage

Open a file directly from the command line:
```bash
md-editor.exe path/to/your/file.md
```

## ⌨️ Keyboard Shortcuts

| Action | Shortcut |
| --- | --- |
| New file | Ctrl + N |
| Open file | Ctrl + O |
| Save | Ctrl + S |
| Save As | Ctrl + Shift + S |
| Bold | Ctrl + B |
| Italic | Ctrl + I |
| Find | Ctrl + F |
| Cycle view (Edit/Split/Preview) | Ctrl + \ |
| Toggle sidebar | Ctrl + Shift + B |
| Zoom in | Ctrl + + or Ctrl + = |
| Zoom out | Ctrl + - |

## 🎨 Markdown Support

- **Text Formatting** — Bold, italic, strikethrough
- **Code** — Inline code and fenced code blocks with syntax highlighting
- **Lists** — Bullet lists, numbered lists, task lists
- **Tables** — Full Markdown table support
- **Blockquotes** — Nested blockquotes
- **Headings** — H1-H6 headings
- **Links & Images** — Standard Markdown syntax
- **Horizontal Rules** — `---` for separators
- **Smart Lists** — Press Enter to continue, empty line to stop

## 🛠️ Development

### Project Structure

```
md-file-reader/
├── app.go              # Go backend - file operations
├── main.go             # Wails application entry point
├── frontend/
│   ├── index.html      # Main HTML structure
│   ├── src/
│   │   ├── app.css     # Application styles
│   │   └── main.js     # Frontend logic
│   └── wailsjs/        # Auto-generated Wails bindings
└── build/              # Build configuration
```

### Running in Development Mode

```bash
wails dev
```

This will start the application with hot-reload enabled.
## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Wails](https://wails.io/) - Go framework for building desktop apps
- Icons from [Heroicons](https://heroicons.com/)

## 📄 Changelog
### v1.5.0 — Windows Installer & Renamed to Markit
- 📦 Windows installer — proper setup.exe with install wizard
- 🔗 .md and .markdown file association — appears in right-click Open with
- 🖥️ Start Menu and Desktop shortcuts created automatically
- 🗑️ Clean uninstaller in Add/Remove Programs
- ✏️ App renamed from MD Editor to Markit
- 📱 Both portable .exe and installer available

### v1.4.0 — File Watching & Conflict Resolution
- 👁️ Automatic file watching — detects when files are modified externally
- ⚠️ Conflict notification bar — appears when external changes conflict with unsaved edits
- 🔄 Reload option — reload file from disk, discarding local changes
- 💾 Save option — overwrite external changes with your version
- 🛡️ Keep option — ignore external changes and continue editing
- 🎯 Smart conflict handling — only shows notification when you have unsaved changes
- 📦 fsnotify integration — efficient file system monitoring via Go backend

### v1.3.0 — Document Outline Panel
- 📋 Document outline panel in sidebar — reads all headings live
- 🖱️ Click any outline item to scroll preview to that heading
- ↕️ Resizable divider between Files and Outline sections
- 🔄 Outline updates live as you type

### v1.2.0 — Zoom Controls and UI Icon Improvements
- Added zoom in/out buttons to the titlebar
- Added keyboard shortcuts for zoom (Ctrl +/-)
- Replaced text buttons with professional SVG icons
- Improved HTML structure for better code readability

### v1.1.0 — Collapsible Sidebar and Startup File Support
- Added collapsible sidebar with smooth animation
- Added persistent sidebar state (saved to localStorage)
- Added support for opening files from command line arguments
- Fixed syntax highlighter for .env and config files
- Improved overall UI polish

### v1.0.2 — File Association and Icon Update
- Open .md files directly — right-click any .md file and open with Markit
- Fixed syntax highlighter breaking .env and config file rendering
- Updated welcome screen
- Updated app icon

### v1.0.1 — Icon Update
- Updated app icon

### v1.0.0 — Initial Release
- Core Markdown editing functionality
- Split view mode
- Native Windows file dialogs
- Keyboard shortcuts
- Syntax highlighting
- Multiple files open at once with sidebar
- Formatting toolbar (Bold, Italic, Headings, Lists, Tables, Code blocks)
- Find in document with match count
- Task lists with clickable checkboxes
- Dark theme

---

> **Tip:** Drag the center divider left or right to resize the panes in Split mode.

