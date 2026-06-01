package generator

import (
	"context"
	"strings"
	"testing"
)

func TestFileNameGenerator(t *testing.T) {
	g := &FileNameGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		name := v.(string)
		if !strings.Contains(name, ".") {
			t.Errorf("expected file extension, got %q", name)
		}
		parts := strings.SplitN(name, ".", 2)
		if len(parts) != 2 {
			t.Errorf("expected name.ext format, got %q", name)
		}
		if len(parts[0]) == 0 {
			t.Errorf("empty file name: %q", name)
		}
	}
}

func TestImageURLGenerator(t *testing.T) {
	g := &ImageURLGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		url := v.(string)
		if !strings.HasPrefix(url, "https://") {
			t.Errorf("expected https URL, got %q", url)
		}
		if !strings.Contains(url, "picsum.photos") {
			t.Errorf("expected picsum.photos URL, got %q", url)
		}
	}
}
