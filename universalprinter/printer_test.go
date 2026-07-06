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
