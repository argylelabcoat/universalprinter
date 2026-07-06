package universalprinter

import (
	"bytes"
	"os"
	"testing"
)

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
		path      string
		wantMime  string
		wantText  bool
		wantPrint bool
	}{
		{"document.pdf", "application/pdf", false, true},
		{"notes.txt", "text/plain", true, true},
		{"image.jpg", "image/jpeg", false, true},
		{"style.html", "text/html", true, true},
		{"data.csv", "text/csv", true, true},
		{"report.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", false, true},
		{"unknown.xyz", "chemical/x-xyz", false, false},
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
