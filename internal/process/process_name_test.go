package process

import "testing"

func TestShortProcessName(t *testing.T) {
	cases := map[string]string{
		"chrome": "chrome",
		"chrome-headless-shell --type=renderer --headless=old --no-sandbox": "chrome-headless-shell",
		"/usr/lib/chromium/chrome --type=renderer":                          "chrome",
		"  node --inspect server.js ":                                       "node",
		"":                                                                  "",
	}
	for in, want := range cases {
		if got := shortProcessName(in); got != want {
			t.Errorf("shortProcessName(%q) = %q, want %q", in, got, want)
		}
	}
}
