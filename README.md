# LogView

LogView is a native Linux (and cross-platform) log file viewer built with Go and the Fyne GUI toolkit. It provides a modern, user-friendly interface for viewing, filtering, and analyzing log files.
Note: This app was generated fully with AI. No code was handwritten.

## Features
- Fast log file loading and indexing
- Powerful filtering and search capabilities
- Tabbed interface for multiple logs
- Customizable plugins and format support
- Cross-platform: Linux, macOS, Windows

## Installation

### From Source
1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/log-view.git
   cd log-view
   ```
2. Build the application:
   ```sh
   ./build.sh
   ```
3. (Optional) Install dependencies:
   ```sh
   go mod tidy
   ```

## Usage
- Open log files via the UI or command line.
- Use the filter bar to search and filter log entries.
- Switch between multiple logs using tabs.

## Development
- Main application: `main.go`
- Core logic: `internal/app/`
- UI components: `internal/ui/`
- Plugins: `internal/plugin/`

### Build
```sh
./build.sh
```

### Run
```sh
go run main.go
```

---

**LogView** — Modern, fast, and extensible log viewer for developers and sysadmins.
