package universalprinter

import (
	"bytes"
	"os"
	"testing"
)

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
		t.Error("PDF missing EOF marker")
	}
}

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
