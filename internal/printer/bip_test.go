package printer

import (
	"context"
	"testing"
)

func TestBip(t *testing.T) {
	ctx := context.Background()
	p := New(ctx)
	err := p.PrintBip()
	if err != nil {
		t.Errorf("Error printing bip: %v", err)
	}
}
