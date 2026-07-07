// Package universalprinter provides cross-platform document printing with
// automatic PDF fallback. It uses only the Go standard library and supports
// macOS, Linux (CUPS), and Windows.
package universalprinter

import (
	"mime"
	"strings"
)

// printableTypes is the set of file extensions the library can handle for
// direct printing via the OS print subsystem.
var printableTypes = map[string]bool{
	".txt": true, ".pdf": true, ".doc": true, ".docx": true,
	".rtf": true, ".odt": true, ".jpg": true, ".jpeg": true,
	".png": true, ".gif": true, ".bmp": true, ".tiff": true,
	".html": true, ".htm": true, ".xml": true, ".csv": true,
	".json": true,
}

// textTypes is the set of file extensions treated as plain text. These can be
// read as strings and written to temp files for printing.
var textTypes = map[string]bool{
	".txt": true, ".csv": true, ".json": true,
	".xml": true, ".html": true, ".htm": true,
}

// isTextFile reports whether the given file extension (e.g. ".txt") represents
// a text-based file type.
func isTextFile(ext string) bool {
	return textTypes[strings.ToLower(ext)]
}

// isPrintable reports whether the given file extension (e.g. ".pdf") is in the
// set of types the library can send directly to the OS printer.
func isPrintable(ext string) bool {
	return printableTypes[strings.ToLower(ext)]
}

// detectFileType determines the MIME type, whether the content is text, and
// whether the file can be printed directly, based on the file extension.
// Falls back to "application/octet-stream" for unknown extensions.
func detectFileType(filePath string) (mimeType string, isText bool, isPrintable bool) {
	ext := ""
	if i := strings.LastIndex(filePath, "."); i != -1 {
		ext = strings.ToLower(filePath[i:])
	}

	mimeType = mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if i := strings.Index(mimeType, ";"); i != -1 {
		mimeType = mimeType[:i]
	}

	isText = textTypes[ext] || strings.HasPrefix(mimeType, "text/")
	isPrintable = printableTypes[ext]

	return mimeType, isText, isPrintable
}
