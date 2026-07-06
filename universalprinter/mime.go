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
	if i := strings.Index(mimeType, ";"); i != -1 {
		mimeType = mimeType[:i]
	}

	isText = textTypes[ext] || strings.HasPrefix(mimeType, "text/")
	isPrintable = printableTypes[ext]

	return mimeType, isText, isPrintable
}
