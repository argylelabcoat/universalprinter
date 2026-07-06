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
		t.Run(tt.ext, func(t *testing.T) {
			if got := isTextFile(tt.ext); got != tt.want {
				t.Errorf("isTextFile(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
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
		t.Run(tt.ext, func(t *testing.T) {
			if got := isPrintable(tt.ext); got != tt.want {
				t.Errorf("isPrintable(%q) = %v, want %v", tt.ext, got, tt.want)
			}
		})
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
		{"unknown.xyz", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			mime, text, print := detectFileType(tt.path)
			if tt.wantMime == "" {
				if mime == "" {
					t.Errorf("detectFileType(%q) mime is empty, want non-empty", tt.path)
				}
			} else {
				if mime != tt.wantMime {
					t.Errorf("detectFileType(%q) mime = %q, want %q", tt.path, mime, tt.wantMime)
				}
			}
			if text != tt.wantText {
				t.Errorf("detectFileType(%q) text = %v, want %v", tt.path, text, tt.wantText)
			}
			if print != tt.wantPrint {
				t.Errorf("detectFileType(%q) print = %v, want %v", tt.path, print, tt.wantPrint)
			}
		})
	}
}
