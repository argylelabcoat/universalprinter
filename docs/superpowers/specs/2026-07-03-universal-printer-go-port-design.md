# Universal Printer — Go Port Design

**Source:** [sharathkumardaroor/universal_printer](https://github.com/sharathkumardaroor/universal_printer) (Python 2.0.0)
**Date:** 2026-07-03

## Overview

Port the Python `universal_printer` library to a minimal, idiomatic Go package. Zero external dependencies. PDF fallback uses raw PDF byte generation (no PDF library). Optional PDF library bindings can come later.

## Package Structure

```
universalprinter/
├── printer.go          # DocumentPrinter struct + public API
├── printer_unix.go     # lp-based printing (darwin, linux)
├── printer_windows.go  # rundll32-based printing
├── pdf.go              # Minimal PDF fallback writer
├── mime.go             # File type detection helpers
├── printer_test.go     # Basic tests
├── go.mod
└── README.md
```

## Public API

### Types

```go
type PrintResult struct {
    Success bool
    Message string
    PDFPath string // empty if no fallback generated
}

type printOptions struct {
    printerName string
    fallback    bool
    pdfFilename string
}
```

### Functional Options

```go
type Option func(*printOptions)

func WithPrinter(name string) Option    // set target printer
func WithFallback(enable bool) Option   // enable/disable PDF fallback (default: true)
func WithPDFFilename(name string) Option // custom PDF fallback filename
```

### DocumentPrinter

```go
type DocumentPrinter struct {
    system   string              // runtime.GOOS
    downloads string             // ~/Downloads path
    printable map[string]bool    // extensions that can be printed directly
}

func New() *DocumentPrinter
```

### Methods

```go
func (p *DocumentPrinter) PrintDocument(contentOrPath string, opts ...Option) (*PrintResult, error)
func (p *DocumentPrinter) PrintText(text string, opts ...Option) (*PrintResult, error)
func (p *DocumentPrinter) PrintFile(filePath string, opts ...Option) (*PrintResult, error)
func (p *DocumentPrinter) GetSupportedFileTypes() []string
func (p *DocumentPrinter) IsFilePrintable(filePath string) bool
```

### Supported File Types

Same as Python original:
`.txt`, `.pdf`, `.doc`, `.docx`, `.rtf`, `.odt`,
`.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.tiff`,
`.html`, `.htm`, `.xml`, `.csv`, `.json`

## File Details

### printer.go (shared logic)

- `New()`: initializes system, downloads path, printable types map
- `PrintDocument()`: main entry point — detects if input is a file path (via `os.Stat`) or text content, dispatches to OS-specific print, falls back to PDF on failure. Default options: fallback=true, no printer name
- `PrintText()` / `PrintFile()`: thin wrappers around `PrintDocument()`
- File type detection via `mime.go`
- Temp file creation for text content, cleanup in `defer`

### printer_unix.go

```go
//go:build darwin || linux

func printFile(filePath, printerName string) error
```

- Runs `lp` command with optional `-d printerName`
- Returns error on failure (non-zero exit)

### printer_windows.go

```go
//go:build windows

func printFile(filePath, printerName string) error
```

- Runs `rundll32.exe shell32.dll,ShellExec_RunDLL <file> print`
- Falls back to `notepad /P` for `.txt` files if needed

### pdf.go

```go
func writeMinimalPDF(content string, outputPath string) error
```

- Raw PDF 1.4 byte generation (same approach as Python)
- Single page, Helvetica font, ASCII text
- Builds PDF objects, xref table, trailer manually
- Writes to `~/Downloads/<filename>`

### mime.go

```go
func detectFileType(filePath string) (mimeType string, isText bool, isPrintable bool)
```

- Extension-based detection (Go's `mime` package supplements)
- Returns mime type, whether it's text, whether it's in the printable set

## Error Handling

- `PrintDocument` returns `(*PrintResult, error)` — `error` for OS-level failures, `PrintResult.Message` for informational details
- PDF fallback is attempted when `fallback` option is true (default)
- If both printing and PDF fallback fail, returns error

## Testing

- Unit tests for `IsFilePrintable`, `GetSupportedFileTypes`, `writeMinimalPDF`
- Integration tests skipped by default (require real printer / OS)
- Table-driven tests for file type detection
