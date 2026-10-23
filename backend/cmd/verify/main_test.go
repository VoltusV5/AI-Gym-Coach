package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerify проверяет запуск каждой цели, остановку при ошибке и сохранение диагностики.
func TestVerify(t *testing.T) {
	for _, failedTarget := range []string{"", "fmt-check", "vet", "test", "build"} {
		name := failedTarget
		if name == "" {
			name = "success"
		}
		t.Run(name, func(t *testing.T) {
			workDir := t.TempDir()
			t.Chdir(workDir)
			var makefile strings.Builder
			targets := []string{"fmt-check", "vet", "test", "build"}
			for _, target := range targets {
				fmt.Fprintf(&makefile, ".PHONY: %s\n%s:\n\t@echo ran-%s\n", target, target, target)
				if target == failedTarget {
					makefile.WriteString("\t@echo 'example.go:12: concrete failure' >&2\n\t@exit 1\n")
				}
			}
			require.NoError(t, os.WriteFile("Makefile", []byte(makefile.String()), 0600))
			logDir := t.TempDir()
			var output bytes.Buffer

			err := verify(&output, logDir)

			if failedTarget == "" {
				require.NoError(t, err)
				assert.NotContains(t, output.String(), "ran-", "successful command output belongs in logs")
			} else {
				require.ErrorContains(t, err, failedTarget)
				assert.Contains(t, output.String(), "FAIL "+failedTarget)
				assert.Contains(t, output.String(), "example.go:12: concrete failure")
				assert.NotContains(t, output.String(), "PASS "+failedTarget)
			}
			for i, target := range targets {
				data, readErr := os.ReadFile(filepath.Join(logDir, target+".log"))
				require.NoError(t, readErr)
				assert.Contains(t, string(data), "ran-"+target)
				if target == failedTarget {
					assert.Contains(t, string(data), "example.go:12: concrete failure")
					for _, skipped := range targets[i+1:] {
						_, statErr := os.Stat(filepath.Join(logDir, skipped+".log"))
						assert.True(t, os.IsNotExist(statErr), "must stop before %s", skipped)
					}
					break
				}
			}
		})
	}
}

// TestFmtCheck проверяет действующий Makefile на чистом, неформатированном и синтаксически неверном Go.
// Ни один из сценариев не должен изменять исходник.
func TestFmtCheck(t *testing.T) {
	makefile, err := os.ReadFile("../../Makefile")
	require.NoError(t, err)
	version, err := os.ReadFile("../../.golangci-version")
	require.NoError(t, err)
	for _, tt := range []struct {
		name   string
		source string
		valid  bool
	}{
		{name: "formatted", source: "package example\n\nvar Value = 1\n", valid: true},
		{name: "unformatted", source: "package example\nvar Value=1\n"},
		{name: "syntax error", source: "package example\nfunc broken(\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "Makefile"), makefile, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, ".golangci-version"), version, 0600))
			path := filepath.Join(dir, "example.go")
			require.NoError(t, os.WriteFile(path, []byte(tt.source), 0600))
			cmd := exec.Command("make", "--no-print-directory", "-s", "fmt-check")
			cmd.Dir = dir

			output, err := cmd.CombinedOutput()

			if tt.valid {
				require.NoError(t, err, "%s", output)
			} else {
				require.Error(t, err)
				assert.Contains(t, string(output), "example.go")
			}
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.source, string(data), "fmt-check must never rewrite sources")
		})
	}
}
