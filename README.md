# universalprinter

A cross-platform document printing library for Go with automatic PDF fallback. Zero external dependencies.

## Features

- **Universal file support** — print any file type (PDF, DOC, TXT, images, HTML, etc.)
- **Cross-platform** — macOS, Linux (CUPS `lp`), Windows (`rundll32.exe`)
- **PDF fallback** — automatically generates a minimal PDF when printing fails
- **File type detection** — MIME type lookup via file extension
- **Zero dependencies** — uses only the Go standard library

## Install

```bash
go get github.com/argylelabcoat/universalprinter
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/argylelabcoat/universalprinter"
)

func main() {
	printer := universalprinter.New()

	// Print text
	result, err := printer.PrintText("Hello, World!\nThis is a test document.")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result.Message)

	// Print a file
	result, err = printer.PrintFile("/path/to/document.pdf")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(result.Message)

	// Print to a specific printer
	result, err = printer.PrintText("Important memo",
		universalprinter.WithPrinter("HP_LaserJet_Pro"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Disable PDF fallback
	result, err = printer.PrintText("Print or fail",
		universalprinter.WithFallback(false),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Check if a file type is supported
	if printer.IsFilePrintable("/path/to/image.jpg") {
		fmt.Println("Image files are directly printable")
	}

	// Get all supported types
	types := printer.GetSupportedFileTypes()
	fmt.Println("Supported types:", types)
}
```

## API

### `New() *DocumentPrinter`

Creates a new printer instance. Detects the current OS and locates the user's Downloads directory for PDF fallback.

### `PrintText(text string, opts ...Option) (*PrintResult, error)`

Prints a text string. Writes to a temp file, sends to the OS print subsystem, and cleans up.

### `PrintFile(filePath string, opts ...Option) (*PrintResult, error)`

Prints an existing file. Returns an error result (not a Go error) if the file doesn't exist.

### `PrintDocument(contentOrPath string, opts ...Option) (*PrintResult, error)`

Universal method — detects whether the input is a file path or text content and dispatches accordingly.

### `GetSupportedFileTypes() []string`

Returns the list of file extensions the library can handle for direct printing.

### `IsFilePrintable(filePath string) bool`

Returns whether the given file's extension is in the supported set.

### Options

| Option | Description |
|--------|-------------|
| `WithPrinter(name string)` | Target a specific printer by name |
| `WithFallback(enable bool)` | Enable/disable PDF fallback (default: true) |
| `WithPDFFilename(name string)` | Custom filename for the PDF fallback |

### `PrintResult`

```go
type PrintResult struct {
    Success bool    // whether the print succeeded
    Message string  // human-readable status
    PDFPath string  // path to fallback PDF (empty if none)
}
```

## Supported File Types

`.txt`, `.pdf`, `.doc`, `.docx`, `.rtf`, `.odt`, `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.tiff`, `.html`, `.htm`, `.xml`, `.csv`, `.json`

## Platform Behavior

| OS | Print command | Fallback |
|----|---------------|----------|
| macOS | `lp` (CUPS) | PDF generation |
| Linux | `lp` (CUPS) | PDF generation |
| Windows | `rundll32.exe shell32.dll,ShellExec_RunDLL` | `notepad /P` for .txt, then PDF generation |

## License

MIT
