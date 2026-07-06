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
	textOps.WriteString("/F1 12 Tf 50 750 Td\n")
	for _, line := range lines {
		safe := strings.ReplaceAll(line, "(", "\\(")
		safe = strings.ReplaceAll(safe, ")", "\\)")
		textOps.WriteString(fmt.Sprintf("(%s) Tj 0 -14 Td\n", safe))
	}

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
