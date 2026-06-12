package state

import (
	"encoding/json"
	"testing"
)

func TestNormalizeRURFromExportedJSON(t *testing.T) {
	raw := `{"cookies":{"rur":"\"CCO\\05474808498212\\0541812777836:abc\""}}`
	var snap map[string]any
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		t.Fatal(err)
	}
	cookies := snap["cookies"].(map[string]any)
	got := normalizeRUR(cookies["rur"].(string))
	want := "CCO,74808498212,1812777836:abc"
	if got != want {
		t.Fatalf("normalizeRUR = %q, want %q", got, want)
	}
}