# Universal Printer Go Port — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port the Python `universal_printer` library to an idiomatic Go package with zero external dependencies.

**Architecture:** Single Go package `universalprinter` with build-tagged OS files, a raw PDF fallback writer, and extension-based file type detection. Functional options pattern for configuration.

**Tech Stack:** Go stdlib only — `os`, `os/exec`, `mime`, `path/filepath`, `runtime`, `testing`.

---

## File Structure

```
universalprinter/
├── go.mod                  # module universalprinter
├── mime.go                 # detectFileType, printableTypes, isTextFile
├── pdf.go                  # writeMinimalPDF
├── printer_unix.go         # printFile for darwin || linux
├── printer_windows.go      # printFile for windows
├── printer.go              # DocumentPrinter struct + public API
├── printer_test.go         # unit tests
└── README.md               # usage docs
```

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`

- [ ] **Step 1: Initialize Go module**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go mod init universalprinter
```

Expected: `go.mod` created with `module universalprinter` and go version.

- [ ] **Step 2: Verify**

```bash
cat go.mod
```

Expected output (approx):
```
module universalprinter

go 1.21
```

- [ ] **Step 3: Commit**

```bash
git add go.mod
git commit -m "feat: initialize go module"
```

---

### Task 2: File Type Detection (mime.go)

**Files:**
- Create: `universalprinter/mime.go`
- Test: `universalprinter/printer_test.go`

- [ ] **Step 1: Write failing tests for file type detection**

Create `universalprinter/printer_test.go`:

```go
package universalprinter

import "testing"

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".txt", true},
		{".csv", true},
		{".json", true},
		{".xml", true},
		{".html", true},
		{".htm", true},
		{".pdf", false},
		{".jpg", false},
		{".docx", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isTextFile(tt.ext); got != tt.want {
			t.Errorf("isTextFile(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestIsPrintable(t *testing.T) {
	tests := []struct {
		ext  string
		want bool
	}{
		{".txt", true},
		{".pdf", true},
		{".doc", true},
		{".docx", true},
		{".jpg", true},
		{".png", true},
		{".html", true},
		{".xyz", false},
		{".bin", false},
	}
	for _, tt := range tests {
		if got := isPrintable(tt.ext); got != tt.want {
			t.Errorf("isPrintable(%q) = %v, want %v", tt.ext, got, tt.want)
		}
	}
}

func TestDetectFileType(t *testing.T) {
	tests := []struct {
		path         string
		wantMime     string
		wantText     bool
		wantPrint    bool
	}{
		{"document.pdf", "application/pdf", false, true},
		{"notes.txt", "text/plain", true, true},
		{"image.jpg", "image/jpeg", false, true},
		{"style.html", "text/html", true, true},
		{"data.csv", "text/csv", true, true},
		{"report.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false, true},
		{"unknown.xyz", "application/octet-stream", false, false},
	}
	for _, tt := range tests {
		mime, text, print := detectFileType(tt.path)
		if mime != tt.wantMime {
			t.Errorf("detectFileType(%q) mime = %q, want %q", tt.path, mime, tt.wantMime)
		}
		if text != tt.wantText {
			t.Errorf("detectFileType(%q) text = %v, want %v", tt.path, text, tt.wantText)
		}
		if print != tt.wantPrint {
			t.Errorf("detectFileType(%q) print = %v, want %v", tt.path, print, tt.wantPrint)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -run "TestIsTextFile|TestIsPrintable|TestDetectFileType" -v
```

Expected: FAIL — `undefined: isTextFile`, `undefined: isPrintable`, `undefined: detectFileType`

- [ ] **Step 3: Implement mime.go**

Create `universalprinter/mime.go`:

```go
package universalprinter

import (
	"mime"
	"strings"
)

var printableTypes = map[string]bool{
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".rtf": true, ".odt": true, ".jpg": true, ".jpeg": true,
	".png": true, ".gif": true, ".bmp": true, ".tiff": true,
	".html": true, ".htm": true, ".xml": true, ".csv": true,
	".json": true,
}

var textTypes = map[string]bool{
	".txt": true, ".csv": true, ".json": true,
	".xml": true, ".html": true, ".htm": true,
}

func isTextFile(ext string) bool {
	return textTypes[strings.ToLower(ext)]
}

func isPrintable(ext string) bool {
	return printableTypes[strings.ToLower(ext)]
}

func detectFileType(filePath string) (mimeType string, isText bool, isPrintable bool) {
	ext := ""
	if i := strings.LastIndex(filePath, "."); i != -1 {
		ext = strings.ToLower(filePath[i:])
	}

	mimeType = mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	isText = textTypes[ext] || strings.HasPrefix(mimeType, "text/")
	isPrint = printableTypes[ext]

	return mimeType, isText, isPrint
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -run "TestIsTextFile|TestIsPrintable|TestDetectFileType" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add universalprinter/mime.go universalprinter/printer_test.go
git commit -m "feat: add file type detection with tests"
```

---

### Task 3: PDF Fallback Writer (pdf.go)

**Files:**
- Create: `universalprinter/pdf.go`
- Test: `universalprinter/printer_test.go` (append)

- [ ] **Step 1: Write failing test for PDF generation**

Append to `universalprinter/printer_test.go`:

```go
func TestWriteMinimalPDF(t *testing.T) {
	path := t.TempDir() + "/test.pdf"
	err := writeMinimalPDF("Hello World\nLine 2", path)
	if err != nil {
		t.Fatalf("writeMinimalPDF returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read generated PDF: %v", err)
	}

	if !bytes.HasPrefix(data, []byte("%PDF-1.4")) {
		t.Errorf("PDF does not start with %%PDF-1.4 header, got: %s", data[:20])
	}

	if !bytes.Contains(data, []byte("Hello World")) {
		t.Error("PDF content does not contain expected text")
	}

	if !bytes.Contains(data, []byte("xref")) {
		t.Error("PDF missing xref table")
	}

	if !bytes.Contains(data, []byte("%%EOF")) {
		t.Error("PDF missing %%EOF marker")
	}
}

func TestWriteMinimalPDF_EmptyContent(t *testing.T) {
	path := t.TempDir() + "/empty.pdf"
	err := writeMinimalPDF("", path)
	if err != nil {
		t.Fatalf("writeMinimalPDF returned error for empty content: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read generated PDF: %v", err)
	}

	if !bytes.HasPrefix(data, []byte("%PDF-1.4")) {
		t.Errorf("PDF does not start with %%PDF-1.4 header")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -run "TestWriteMinimalPDF" -v
```

Expected: FAIL — `undefined: writeMinimalPDF`

- [ ] **Step 3: Implement pdf.go**

Create `universalprinter/pdf.go`:

```go
package universalprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func writeMinimalPDF(content string, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	lines := strings.Split(content, "\n")

	var pdf []byte
	var offsets []int

	addObject := func(obj string) int {
		offsets = append(offsets, len(pdf))
		pdf = append(pdf, []byte(fmt.Sprintf("%d 0 obj\n", len(offsets)))...)
		pdf = append(pdf, []byte(obj)...)
		pdf = append(pdf, []byte("\nendobj\n")...)
		return len(offsets)
	}

	pdf = append(pdf, []byte("%PDF-1.4\n")...)

	fontNum := addObject("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	var textOps strings.Builder
	textOps.WriteString("BT /F1 12 Tf 50 750 Td\n")
	for _, line := range lines {
		safe := strings.ReplaceAll(line, "(", "\\(")
		safe = strings.ReplaceAll(safe, ")", "\\)")
		textOps.WriteString(fmt.Sprintf("(%s) Tj 0 -14 Td\n", safe))
	}
	textOps.WriteString("ET")

	stream := fmt.Sprintf("q\n1 0 0 1 0 0 cm\nBT\n/F1 12 Tf\n50 750 Td\n%s\nET\nQ", textOps.String())
	streamBytes := []byte(stream)

	contentNum := addObject(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamBytes), stream))

	pageNum := addObject(fmt.Sprintf(
		"<< /Type /Page /Parent 4 0 R /Resources << /Font << /F1 %d 0 R >> >> /Contents %d 0 R /MediaBox [0 0 612 792] >>",
		fontNum, contentNum,
	))

	pagesNum := addObject(fmt.Sprintf("<< /Type /Pages /Kids [ %d 0 R ] /Count 1 >>", pageNum))

	catalogNum := addObject(fmt.Sprintf("<< /Type /Catalog /Pages %d 0 R >>", pagesNum))

	xrefStart := len(pdf)
	pdf = append(pdf, []byte("xref\n")...)
	pdf = append(pdf, []byte(fmt.Sprintf("0 %d\n", len(offsets)+1))...)
	pdf = append(pdf, []byte("0000000000 65535 f \n")...)
	for _, off := range offsets {
		pdf = append(pdf, []byte(fmt.Sprintf("%010d 00000 n \n", off))...)
	}

	pdf = append(pdf, []byte("trailer\n")...)
	pdf = append(pdf, []byte(fmt.Sprintf("<< /Size %d /Root %d 0 R >>\n", len(offsets)+1, catalogNum))...)
	pdf = append(pdf, []byte("startxref\n")...)
	pdf = append(pdf, []byte(fmt.Sprintf("%d\n", xrefStart))...)
	pdf = append(pdf, []byte("%%EOF\n")...)

	return os.WriteFile(outputPath, pdf, 0644)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -run "TestWriteMinimalPDF" -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add universalprinter/pdf.go
git commit -m "feat: add minimal PDF fallback writer"
```

---

### Task 4: OS-Specific Print Functions

**Files:**
- Create: `universalprinter/printer_unix.go`
- Create: `universalprinter/printer_windows.go`

- [ ] **Step 1: Implement printer_unix.go**

Create `universalprinter/printer_unix.go`:

```go
//go:build darwin || linux

package universalprinter

import (
	"fmt"
	"os/exec"
)

func printFile(filePath string, printerName string) error {
	args := []string{}
	if printerName != "" {
		args = append(args, "-d", printerName)
	}
	args = append(args, filePath)

	cmd := exec.Command("lp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lp failed: %w, output: %s", err, string(output))
	}
	return nil
}
```

- [ ] **Step 2: Implement printer_windows.go**

Create `universalprinter/printer_windows.go`:

```go
//go:build windows

package universalprinter

import (
	"fmt"
	"os/exec"
	"strings"
)

func printFile(filePath string, printerName string) error {
	cmd := exec.Command("rundll32.exe", "shell32.dll,ShellExec_RunDLL", filePath, "print")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to notepad for .txt files
		if strings.HasSuffix(strings.ToLower(filePath), ".txt") {
			cmd = exec.Command("notepad", "/P", filePath)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("notepad print failed: %w, output: %s", err, string(output))
			}
			return nil
		}
		return fmt.Errorf("rundll32 print failed: %w, output: %s", err, string(output))
	}
	return nil
}
```

- [ ] **Step 3: Verify compilation (no runtime test needed)**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
GOOS=darwin go build ./...
GOOS=linux go build ./...
GOOS=windows go build ./...
```

Expected: All compile without error.

- [ ] **Step 4: Commit**

```bash
git add universalprinter/printer_unix.go universalprinter/printer_windows.go
git commit -m "feat: add OS-specific print functions with build tags"
```

---

### Task 5: Main API (printer.go)

**Files:**
- Create: `universalprinter/printer.go`
- Test: `universalprinter/printer_test.go` (append)

- [ ] **Step 1: Write failing tests for DocumentPrinter**

Append to `universalprinter/printer_test.go`:

```go
func TestNew(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.system == "" {
		t.Error("system not set")
	}
	if p.downloads == "" {
		t.Error("downloads path not set")
	}
}

func TestGetSupportedFileTypes(t *testing.T) {
	p := New()
	types := p.GetSupportedFileTypes()
	if len(types) == 0 {
		t.Error("GetSupportedFileTypes returned empty set")
	}
	for _, ext := range []string{".txt", ".pdf", ".jpg", ".html"} {
		found := false
		for _, got := range types {
			if got == ext {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GetSupportedFileTypes missing %s", ext)
		}
	}
}

func TestIsFilePrintable(t *testing.T) {
	p := New()
	tests := []struct {
		path string
		want bool
	}{
		{"doc.pdf", true},
		{"image.jpg", true},
		{"unknown.xyz", false},
	}
	for _, tt := range tests {
		if got := p.IsFilePrintable(tt.path); got != tt.want {
			t.Errorf("IsFilePrintable(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestPrintText_NoFallback(t *testing.T) {
	p := New()
	result, err := p.PrintText("test", WithFallback(false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Printing may fail if no printer configured, but should not error
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestPrintFile_NotFound(t *testing.T) {
	p := New()
	result, err := p.PrintFile("/nonexistent/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected Success=false for nonexistent file")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -run "TestNew|TestGetSupported|TestIsFilePrintable|TestPrintText_NoFallback|TestPrintFile_NotFound" -v
```

Expected: FAIL — `undefined: New`, `undefined: DocumentPrinter`, etc.

- [ ] **Step 3: Implement printer.go**

Create `universalprinter/printer.go`:

```go
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

	// Detect if input is a file path
	if info, err := os.Stat(contentOrPath); err == nil && !info.IsDir() {
		return p.printFilePath(contentOrPath, o)
	}

	// Treat as text content
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
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -v
```

Expected: All tests PASS

- [ ] **Step 5: Commit**

```bash
git add universalprinter/printer.go
git commit -m "feat: implement DocumentPrinter public API"
```

---

### Task 6: Final Verification

- [ ] **Step 1: Run all tests**

```bash
cd /Volumes/ExternalRAID/Users/matthew/Projects/universal_printer/universalprinter
go test -v -count=1
```

Expected: All PASS

- [ ] **Step 2: Run go vet**

```bash
go vet ./...
```

Expected: No issues

- [ ] **Step 3: Build for all platforms**

```bash
GOOS=darwin GOARCH=arm64 go build ./...
GOOS=linux GOARCH=amd64 go build ./...
GOOS=windows GOARCH=amd64 go build ./...
```

Expected: All compile

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: universal printer go port complete"
```
