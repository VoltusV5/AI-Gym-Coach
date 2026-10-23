// Command verify запускает существующие проверки Makefile без изменения исходников.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// main сохраняет логи во временный каталог и завершает процесс с кодом 1 при сбое проверки.
func main() {
	logDir, err := os.MkdirTemp("", "sportik-verify-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL verify: create logs: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Logs: %s\n", logDir)
	if err := verify(os.Stdout, logDir); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL verify: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS verify")
}

// verify последовательно выполняет цели Makefile из текущего каталога backend.
// Пишет краткий результат в output, полные логи в logDir; возвращает первый сбой команды или записи лога.
func verify(output io.Writer, logDir string) error {
	for _, target := range []string{"fmt-check", "vet", "test", "build"} {
		cmd := exec.Command("make", "--no-print-directory", "-s", target)
		if target == "test" {
			// Успешные тесты не засоряют диагностику ошибки другого пакета.
			cmd.Args = append(cmd.Args, "TEST_FLAGS=")
		}
		data, runErr := cmd.CombinedOutput()
		logPath := filepath.Join(logDir, target+".log")
		if err := os.WriteFile(logPath, data, 0600); err != nil {
			return fmt.Errorf("write %s log: %w", target, err)
		}
		if runErr != nil {
			_, _ = fmt.Fprintf(output, "FAIL %s (log: %s)\n", target, logPath)
			_, _ = fmt.Fprint(output, string(data))
			return fmt.Errorf("%s: %w", target, runErr)
		}
		_, _ = fmt.Fprintf(output, "PASS %s\n", target)
	}
	return nil
}
