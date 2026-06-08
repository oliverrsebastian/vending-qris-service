package qr

import "testing"

func TestRender(t *testing.T) {
	img, err := Render("000201010212")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if img.Size == 0 {
		t.Fatal("expected non-zero size")
	}
	if len(img.Modules) != img.Size {
		t.Fatalf("rows=%d size=%d", len(img.Modules), img.Size)
	}
	for i, row := range img.Modules {
		if len(row) != img.Size {
			t.Fatalf("row %d width=%d size=%d", i, len(row), img.Size)
		}
	}
	if img.ASCII == "" {
		t.Fatal("expected ascii preview")
	}
}

func TestRenderEmpty(t *testing.T) {
	if _, err := Render(""); err == nil {
		t.Fatal("expected error for empty content")
	}
}
