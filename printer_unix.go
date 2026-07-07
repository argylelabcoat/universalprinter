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
