//go:build darwin || linux

package universalprinter

import (
	"fmt"
	"os/exec"
)

// printFile sends a file to the printer using the CUPS lp command.
// On macOS and Linux, this is the standard way to print from the command line.
// If printerName is non-empty, it is passed as the -d flag to lp.
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
