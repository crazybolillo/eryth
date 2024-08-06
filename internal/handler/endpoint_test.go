package handler

import "testing"

func TestDisplayNameFromClid(t *testing.T) {
	cases := map[string]struct {
		callerID    string
		displayName string
	}{
		"valid": {
			`"Kiwi Snow" <kiwi99>`,
			"Kiwi Snow",
		},
		"empty": {
			"",
			"",
		},
		"single_colon": {
			`"`,
			"",
		},
		"empty_quotes": {
			`""`,
			"",
		},
		"missing_start_quote": {
			`John Smith" <john>`,
			"",
		},
		"missing_end_quote": {
			`"John Smith <john>`,
			"",
		},
	}

	for name, tt := range cases {
		t.Run(name, func(t *testing.T) {
			got := displayNameFromClid(tt.callerID)
			if got != tt.displayName {
				t.Errorf("got %q, want %q", got, tt.displayName)
			}
		})
	}
}
