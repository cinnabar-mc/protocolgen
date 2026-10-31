//go:build linux || darwin

package sourcelock

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDirectoryDigestRejectsNamedPipes(t *testing.T) {
	if os.Getenv("PROTOCOLGEN_FIFO_CHILD") == "1" {
		root := t.TempDir()
		if err := syscall.Mkfifo(filepath.Join(root, "pipe"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := DigestDirectory(root); err == nil || !strings.Contains(err.Error(), "unsupported file type") {
			t.Fatalf("DigestDirectory = %v", err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDirectoryDigestRejectsNamedPipes$")
	cmd.Env = append(os.Environ(), "PROTOCOLGEN_FIFO_CHILD=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("named pipe verification failed: %v (%v)\n%s", err, ctx.Err(), output)
	}
}
