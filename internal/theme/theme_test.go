package theme

import "testing"

func TestPresetsAreDistinct(t *testing.T) {
	presets := Presets()
	if len(presets) < 2 {
		t.Fatalf("esperava ao menos 2 presets, obteve %d", len(presets))
	}
	if presets[0].Name != "default" {
		t.Errorf("primeiro preset = %q, quer default", presets[0].Name)
	}

	seen := make(map[string]bool)
	for _, preset := range presets {
		if seen[preset.Name] {
			t.Errorf("nome de preset duplicado: %q", preset.Name)
		}
		seen[preset.Name] = true
	}
}
