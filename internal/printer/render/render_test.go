package render_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jpporta/ticket-control/internal/printer/render"
)

func TestRenderSourceToImage(t *testing.T) {
	src := `
#set page(width: 300pt, height: auto, margin: 10pt)
#set text(size: 10pt)

= Test Title

This is a test paragraph for rendering.
`
	img, err := render.RenderSourceToImage(src, "test")
	if err != nil {
		t.Fatalf("RenderSourceToImage failed: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	b := img.Bounds()
	if b.Max.Y%8 != 0 {
		t.Errorf("image height %d is not a multiple of 8", b.Max.Y)
	}
}

func TestRenderFileToImage(t *testing.T) {
	tmpDir := t.TempDir()
	typPath := filepath.Join(tmpDir, "test.typ")
	content := `
#set page(width: 300pt, height: auto, margin: 10pt)
#set text(size: 10pt)

= File Test

Testing file compilation.
`
	if err := os.WriteFile(typPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	img, err := render.RenderFileToImage(typPath)
	if err != nil {
		t.Fatalf("RenderFileToImage failed: %v", err)
	}
	if img == nil {
		t.Fatal("expected non-nil image")
	}
	b := img.Bounds()
	if b.Max.Y%8 != 0 {
		t.Errorf("image height %d is not a multiple of 8", b.Max.Y)
	}
}
