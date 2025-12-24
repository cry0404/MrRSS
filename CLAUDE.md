# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MrRSS is a modern, cross-platform desktop RSS reader built with **Wails v3** (Go backend + Vue.js frontend). It prioritizes privacy, user experience, and performance with features like auto-translation, smart feed discovery, and AI-powered summarization.

## Development Commands

### Frontend (Vue.js + TypeScript)

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Development server with hot reload
npm run dev

# Build for production
npm run build

# Lint and fix code
npm run lint

# Format code
npm run format

# Run unit tests
npm test

# Run tests with UI
npm run test:ui

# Run E2E tests
npm run test:e2e
```

### Backend (Go)

```bash
# Install dependencies
go mod download

# Run backend tests
go test ./...

# Run tests with coverage
go test -v -timeout=5m -coverprofile=coverage.out -covermode=atomic ./internal/...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
go test -v ./internal/database

# Lint backend
go vet ./...

# Format backend code
gofmt -w .
goimports -w .
```

### Wails Application

```bash
# Development mode with hot reload (recommended)
wails3 dev

# Or using Task runner
task dev

# Build for current platform
wails3 build
# Or
task build

# Package with installer
task package

# Run the built application
task run

# List all available tasks
task --list
```

### Using Makefile (Cross-platform)

```bash
# Show all available commands
make help

# Full development check (lint + test + build)
make check

# Setup development environment
make setup

# Clean build artifacts
make clean

# Platform-specific builds
make build-windows
make build-linux
make build-darwin
```

### Pre-commit Hooks

```bash
# Install hooks
pre-commit install

# Run on all files
pre-commit run --all-files
```

## Architecture Overview

### High-Level Structure

- **Backend**: Go 1.24+ with Wails v3, SQLite database
- **Frontend**: Vue 3.5+ Composition API, Pinia state management, TypeScript
- **Communication**: HTTP API (primary) + Wails bindings (system integration)
- **Build System**: Wails v3 + Task runner + Vite

### Key Directories

```plaintext
MrRSS/
├── main.go                    # Application entry point
├── internal/                 # Go backend packages
│   ├── database/            # Data layer, models, migrations
│   ├── handlers/            # HTTP API handlers by feature
│   ├── feed/               # RSS fetching and processing
│   ├── translation/        # Multi-language support
│   ├── discovery/          # Feed discovery engine
│   └── utils/              # Shared utilities
├── frontend/                 # Vue.js frontend
│   ├── src/
│   │   ├── components/      # UI components
│   │   ├── stores/          # Pinia state management
│   │   ├── composables/    # Vue composables
│   │   ├── types/           # TypeScript definitions
│   │   └── i18n/            # Internationalization
│   └── dist/                # Built assets (embedded)
└── wails.json               # Wails configuration
```

### Communication Pattern

The application uses a **hybrid approach**:

1. **HTTP API** (`/api/*`) - Primary communication for data operations
2. **Wails Bindings** - System integration (browser, window management)
3. **Static Files** - Frontend assets served from embedded `frontend/dist`

### Database Schema

Core tables:

- `feeds` - RSS subscriptions with metadata
- `articles` - Individual feed items with read/favorite status
- `settings` - Key-value configuration storage
- `translation_cache` - Cached translations for performance

Important: The database uses SQLite with WAL mode for better concurrency.

## Code Patterns and Guidelines

### Backend Patterns

1. **Context Usage**: Always use `context.Context` for exported methods
2. **Error Handling**: Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
3. **Database Operations**: Use prepared statements for all queries
4. **Input Validation**: Validate URLs, file paths, and user inputs
5. **Resource Cleanup**: Use `defer` for proper cleanup

### Frontend Patterns

1. **Vue 3 Composition API**: Use `<script setup>` syntax
2. **State Management**: Access store via `useAppStore()`
3. **Internationalization**: Always use `t()` for user-facing strings
4. **Error Handling**: Show toast notifications with `window.showToast()`
5. **Type Safety**: Use TypeScript with proper type annotations

### Settings Management

When modifying settings, update ALL required locations:

1. Frontend defaults: `config/defaults.json`
2. Backend defaults: `internal/config/defaults.json`
3. Backend types: `internal/config/config.go`
4. Database initialization: `internal/database/db.go`
5. API handlers: `internal/handlers/settings/settings_handlers.go`
6. Frontend types: `frontend/src/types/settings.ts`
7. Frontend composables: `frontend/src/composables/core/useSettings.ts`
8. Auto-save logic: `frontend/src/composables/core/useSettingsAutoSave.ts`
9. UI Components: Add to appropriate settings tab (e.g., `ReadingSettings.vue`)
10. Internationalization: Add translations to `frontend/src/i18n/locales/en.ts` and `zh.ts`

### Adding a New Setting: Complete Checklist

When adding a new setting (example: `auto_show_all_content`), follow these steps:

#### Backend Changes (5 files)

1. **`config/defaults.json`**
   - Add default value: `"auto_show_all_content": false`

2. **`internal/config/defaults.json`**
   - Add default value: `"auto_show_all_content": false`

3. **`internal/config/config.go`**
   - Add field to `Defaults` struct: `AutoShowAllContent bool `json:"auto_show_all_content"`
   - Add case to `GetString()` function: `case "auto_show_all_content": return strconv.FormatBool(defaults.AutoShowAllContent)`

4. **`internal/database/db.go`**
   - Add to `settingsKeys` array: `"auto_show_all_content"` (ensures database initialization)

5. **`internal/handlers/settings/settings_handlers.go`**
   - GET: Add variable: `autoShowAllContent, _ := h.DB.GetSetting("auto_show_all_content")`
   - GET: Add to JSON response: `"auto_show_all_content": autoShowAllContent`
   - POST: Add field to request struct: `AutoShowAllContent string `json:"auto_show_all_content"`
   - POST: Add save logic: `if req.AutoShowAllContent != "" { h.DB.SetSetting("auto_show_all_content", req.AutoShowAllContent) }`

#### Frontend Changes (6 files)

6. **`frontend/src/types/settings.ts`**
   - Add to `SettingsData` interface: `auto_show_all_content: boolean;`

7. **`frontend/src/composables/core/useSettings.ts`**
   - Add to initial settings object: `auto_show_all_content: settingsDefaults.auto_show_all_content`
   - Add to `fetchSettings()` response: `auto_show_all_content: data.auto_show_all_content === 'true'`

8. **`frontend/src/composables/core/useSettingsAutoSave.ts`**
   - Add to save payload: `auto_show_all_content: (settingsRef.value.auto_show_all_content ?? settingsDefaults.auto_show_all_content).toString()`
   - Add event dispatch: `window.dispatchEvent(new CustomEvent('auto-show-all-content-changed', { detail: { value: settingsRef.value.auto_show_all_content } }))`

9. **`frontend/src/i18n/locales/en.ts`**
   - Add translations: `autoShowAllContent: 'Auto Show All Content'` and `autoShowAllContentDesc: '...'`

10. **`frontend/src/i18n/locales/zh.ts`**
    - Add translations: `autoShowAllContent: '自动展示所有内容'` and `autoShowAllContentDesc: '...'`

11. **UI Component** (e.g., `frontend/src/components/modals/settings/general/ReadingSettings.vue`)
    - Add setting UI with checkbox/toggle
    - Emit update event on change

#### Implementation Logic (if needed)

12. **Feature Implementation**
    - Add composable or logic to implement the setting's functionality
    - Listen for setting change events via `window.addEventListener()`
    - Apply setting logic when value changes

#### Common Pitfalls

- **Missing database key**: Forgetting to add to `db.go` settingsKeys array causes setting to not persist
- **Boolean conversion**: Backend returns strings, frontend must convert: `data.setting === 'true'`
- **Type mismatch**: Ensure boolean settings use `.toString()` when sending to backend
- **Event naming**: Use kebab-case for event names (e.g., `auto-show-all-content-changed`)

## Testing

### Backend Tests

- Run tests with timeout: `go test -v -timeout=5m ./...`
- Coverage report: `go test -coverprofile=coverage.out ./...`
- Single test: `go test -v ./internal/database -run TestSpecificFunction`

### Frontend Tests

- Unit tests: `npm test` (uses Vitest)
- E2E tests: `npm run test:e2e` (uses Cypress)
- Test UI: `npm run test:ui`

## Important Notes

1. **No Wails Bindings for Data**: The application primarily uses HTTP API, not Wails bindings for data operations
2. **Privacy-First**: No external analytics, all data stored locally
3. **Cross-Platform**: Build for Windows, macOS, and Linux
4. **Portable Mode**: Supports portable deployment with `portable.txt`
5. **Single Instance**: Enforced on Windows/macOS, disabled on Linux due to D-Bus issues
6. **Concurrent Processing**: Feed fetching uses goroutines with configurable limits

## Common Issues

1. **Linux D-Bus Issues**: Single instance mode disabled on Linux
2. **Build Requirements**: Ensure platform-specific dependencies are installed
3. **Frontend Hot Reload**: Use `wails3 dev` for development with hot reload
4. **Database Migrations**: Handle schema changes carefully with proper versioning
