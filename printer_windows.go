//go:build windows

package universalprinter

import (
	"fmt"
	"os/exec"
	"strings"
)

// printFile sends a file to the printer using the Windows ShellExecute print
// verb via rundll32.exe. If that fails for .txt files, it falls back to
// notepad /P.
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
