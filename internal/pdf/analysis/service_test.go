package analysis

import "testing"

func TestParseMetadataFromRaw(t *testing.T) {
	raw := `/Title (Quarterly Report) /Author (ACME Inc.) /Producer (PDF Engine)`
	meta := parseMetadataFromRaw(raw)
	if meta["title"] != "Quarterly Report" {
		t.Fatalf("unexpected title: %q", meta["title"])
	}
	if meta["author"] != "ACME Inc." {
		t.Fatalf("unexpected author: %q", meta["author"])
	}
}

func TestCountWords(t *testing.T) {
	text := "Hello world. This is a test with 123 tokens."
	if got := countWords(text); got != 9 {
		t.Fatalf("expected 9 words, got %d", got)
	}
}

func TestDetectTableRows(t *testing.T) {
	text := "Item  Qty  Price\nApple  2  3.50\nNotes: test\nA|B|C"
	rows := detectTableRows(text, 10)
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 table-like rows, got %d", len(rows))
	}
}
