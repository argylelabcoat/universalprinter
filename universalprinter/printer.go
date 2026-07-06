package universalprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type PrintResult struct {
	Success bool
	Message string
	PDFPath string
}

type printOptions struct {
	printerName string
	fallback    bool
	pdfFilename string
}

type Option func(*printOptions)

func WithPrinter(name string) Option {
	return func(o *printOptions) { o.printerName = name }
}

func WithFallback(enable bool) Option {
	return func(o *printOptions) { o.fallback = enable }
}

func WithPDFFilename(name string) Option {
	return func(o *printOptions) { o.pdfFilename = name }
}

type DocumentPrinter struct {
	system    string
	downloads string
	printable map[string]bool
}

func New() *DocumentPrinter {
	home, _ := os.UserHomeDir()
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

func (p *DocumentPrinter) GetSupportedFileTypes() []string {
	types := make([]string, 0, len(p.printable))
	for ext := range p.printable {
		types = append(types, ext)
	}
	return types
}

func (p *DocumentPrinter) IsFilePrintable(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return p.printable[ext]
}

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

func (p *DocumentPrinter) PrintText(text string, opts ...Option) (*PrintResult, error) {
	return p.PrintDocument(text, opts...)
}

func (p *DocumentPrinter) PrintFile(filePath string, opts ...Option) (*PrintResult, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &PrintResult{
			Success: false,
			Message: fmt.Sprintf("File not found: %s", filePath),
		}, nil
	}
	return p.PrintDocument(filePath, opts...)
}

func (p *DocumentPrinter) printFilePath(filePath string, o printOptions) (*PrintResult, error) {
	mimeType, _, _ := detectFileType(filePath)

	err := printFile(filePath, o.printerName)
	if err == nil {
		return &PrintResult{
			Success: true,
			Message: fmt.Sprintf("Printed successfully. MIME type: %s", mimeType),
		}, nil
	}

	if o.fallback {
		pdfPath, fallbackErr := p.savePDFFallback(
			fmt.Sprintf("[File: %s]\n[MIME: %s]\n[Size: %d bytes]", filePath, mimeType, fileSize(filePath)),
			o.pdfFilename,
		)
		if fallbackErr == nil {
			return &PrintResult{
				Success: false,
				Message: fmt.Sprintf("Printing failed. PDF fallback saved to: %s", pdfPath),
				PDFPath: pdfPath,
			}, nil
		}
	}

	return &PrintResult{
		Success: false,
		Message: fmt.Sprintf("Printing failed: %v", err),
	}, err
}

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

	if o.fallback {
		pdfPath, fallbackErr := p.savePDFFallback(text, o.pdfFilename)
		if fallbackErr == nil {
			return &PrintResult{
				Success: false,
				Message: fmt.Sprintf("Printing failed. PDF fallback saved to: %s", pdfPath),
				PDFPath: pdfPath,
			}, nil
		}
	}

	return &PrintResult{
		Success: false,
		Message: fmt.Sprintf("Printing failed: %v", err),
	}, err
}

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

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
