package sourcelock

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryDigestRetainsLengthPrefixedFormat(t *testing.T) {
	root := t.TempDir()
	hash := sha256.New()
	for _, entry := range []struct{ name, data string }{{"a.json", "first\n"}, {"nested/b.json", "second\x00value"}} {
		path := filepath.Join(root, entry.name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(entry.data), 0600); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(hash, "%d:%s\n%d:%s\n", len(entry.name), entry.name, len(entry.data), entry.data)
	}
	got, err := DigestDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := "sha256:" + hex.EncodeToString(hash.Sum(nil)); got != want {
		t.Fatalf("digest %s, want %s", got, want)
	}
}

func BenchmarkDigestDirectory(b *testing.B) {
	for _, size := range []int{1 << 20, 16 << 20} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			root := b.TempDir()
			data := make([]byte, size)
			for index := range 8 {
				if err := os.WriteFile(filepath.Join(root, fmt.Sprint(index)), data, 0600); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.SetBytes(int64(8 * size))
			b.ResetTimer()
			for range b.N {
				if _, err := DigestDirectory(root); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
