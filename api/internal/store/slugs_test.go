package store

import "testing"

func TestSlugify(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"english words", "Acme Corporation", "acme-corporation"},
		{"punctuation", " NIST / CSF 2.0 ", "nist-csf-2-0"},
		{"thai text", "บริษัท เอ บี ซี", "บริษัท-เอ-บี-ซี"},
		{"empty punctuation", "---", "item"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Slugify(test.input); got != test.want {
				t.Fatalf("Slugify(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
