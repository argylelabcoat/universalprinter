package universalprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// PrintResult holds the outcome of a print operation.
type PrintResult struct {
	Success bool   // whether the print succeeded
	Message string // human-readable status message
	PDFPath string // path to fallback PDF, empty if none was generated
}

// printOptions holds configuration for a single print call.
type printOptions struct {
	printerName string // target printer name (empty for system default)
	fallback    bool   // whether to generate a PDF if printing fails
	pdfFilename string // custom filename for PDF fallback
}

// Option configures a print operation.
type Option func(*printOptions)

// WithPrinter targets a specific printer by name. On macOS/Linux this maps to
// the lp -d flag. On Windows this is passed to the shell print verb.
func WithPrinter(name string) Option {
	return func(o *printOptions) { o.printerName = name }
}

// WithFallback enables or disables PDF fallback. When enabled (the default),
// a minimal PDF is generated in ~/Downloads if the OS print command fails.
func WithFallback(enable bool) Option {
	return func(o *printOptions) { o.fallback = enable }
}

// WithPDFFilename sets a custom filename for the PDF fallback. If the name
// doesn't end in ".pdf", the extension is appended automatically.
func WithPDFFilename(name string) Option {
	return func(o *printOptions) { o.pdfFilename = name }
}

// DocumentPrinter is the main entry point for printing. Create one with New.
type DocumentPrinter struct {
	system    string          // runtime.GOOS
	downloads string          // ~/Downloads path for PDF fallback
	printable map[string]bool // extensions supported for direct printing
}

// New creates a new DocumentPrinter. It detects the current OS and locates the
// user's Downloads directory for PDF fallback output.
func New() *DocumentPrinter {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	printable := make(map[string]bool, len(printableTypes))
	for k := range printableTypes {
		printable[k] = true
	}
	return &DocumentPrinter{
		system:    runtime.GOOS,
		downloads: filepath.Join(home, "Downloads"),
		printable: printable,
	}
}

// GetSupportedFileTypes returns the list of file extensions the library can
// handle for direct printing.
func (p *DocumentPrinter) GetSupportedFileTypes() []string {
	types := make([]string, 0, len(p.printable))
	for ext := range p.printable {
		types = append(types, ext)
	}
	return types
}

// IsFilePrintable reports whether the given file's extension is in the set of
// types the library can send directly to the OS printer.
func (p *DocumentPrinter) IsFilePrintable(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return p.printable[ext]
}

// PrintDocument is the universal print method. It detects whether contentOrPath
// is a file path or text content, then dispatches to the appropriate handler.
// If printing fails and fallback is enabled, a minimal PDF is generated.
func (p *DocumentPrinter) PrintDocument(contentOrPath string, opts ...Option) (*PrintResult, error) {
	o := printOptions{fallback: true}
	for _, opt := range opts {
		opt(&o)
	}

	if info, err := os.Stat(contentOrPath); err == nil && !info.IsDir() {
		return p.printFilePath(contentOrPath, o)
	}

	return p.printTextContent(contentOrPath, o)
}

// PrintText is a convenience method that prints a text string. The text is
// written to a temp file, sent to the OS printer, and the temp file is cleaned up.
func (p *DocumentPrinter) PrintText(text string, opts ...Option) (*PrintResult, error) {
	return p.PrintDocument(text, opts...)
}

// PrintFile is a convenience method that prints an existing file. Returns an
// error result (not a Go error) if the file does not exist.
func (p *DocumentPrinter) PrintFile(filePath string, opts ...Option) (*PrintResult, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &PrintResult{
			Success: false,
			Message: fmt.Sprintf("File not found: %s", filePath),
		}, nil
	}
	return p.PrintDocument(filePath, opts...)
}

// printFilePath sends an existing file to the OS printer and handles fallback.
func (p *DocumentPrinter) printFilePath(filePath string, o printOptions) (*PrintResult, error) {
	mimeType, _, _ := detectFileType(filePath)

	err := printFile(filePath, o.printerName)
	if err == nil {
		return &PrintResult{
			Success: true,
			Message: fmt.Sprintf("Printed successfully. MIME type: %s", mimeType),
		}, nil
	}

	return p.handleFallback(err, fmt.Sprintf("[File: %s]\n[MIME: %s]\n[Size: %d bytes]", filePath, mimeType, fileSize(filePath)), o)
}

// printTextContent writes text to a temp file, prints it, and cleans up.
func (p *DocumentPrinter) printTextContent(text string, o printOptions) (*PrintResult, error) {
	tmpFile, err := os.CreateTemp("", "universal-printer-*.txt")
	if err != nil {
		return nil, fmt.Errorf("creating temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(text); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	err = printFile(tmpFile.Name(), o.printerName)
	if err == nil {
		return &PrintResult{
			Success: true,
			Message: "Printed successfully.",
		}, nil
	}

	return p.handleFallback(err, text, o)
}

// savePDFFallback generates a minimal PDF in the Downloads directory. If no
// filename is provided, a timestamped name is used.
func (p *DocumentPrinter) savePDFFallback(content string, filename string) (string, error) {
	if filename == "" {
		filename = fmt.Sprintf("document_%s.pdf", time.Now().Format("20060102_150405"))
	} else if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}

	pdfPath := filepath.Join(p.downloads, filename)
	err := writeMinimalPDF(content, pdfPath)
	if err != nil {
		return "", err
	}
	return pdfPath, nil
}

// handleFallback is the common fallback logic for both file and text printing.
// If fallback is disabled, it returns an error result. Otherwise it generates
// a PDF and returns the path.
func (p *DocumentPrinter) handleFallback(err error, content string, o printOptions) (*PrintResult, error) {
	if !o.fallback {
		return &PrintResult{
			Success: false,
			Message: fmt.Sprintf("Printing failed: %v", err),
		}, err
	}

	pdfPath, fallbackErr := p.savePDFFallback(content, o.pdfFilename)
	if fallbackErr != nil {
		return &PrintResult{
			Success: false,
			Message: fmt.Sprintf("Printing failed: %v", err),
		}, err
	}

	return &PrintResult{
		Success: false,
		Message: fmt.Sprintf("Printing failed. PDF fallback saved to: %s", pdfPath),
		PDFPath: pdfPath,
	}, nil
}

// fileSize returns the size of the file at path in bytes, or 0 on error.
func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
