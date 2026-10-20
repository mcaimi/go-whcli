package whcli

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mcaimi/go-whcli/pkg/api"
)

// runCaptured runs fn with stdout/stderr redirected to pipes and returns
// the captured output.
func runCaptured(t *testing.T, fn func()) (string, string) {
	t.Helper()

	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = outW
	os.Stderr = errW
	fn()
	os.Stdout = oldOut
	os.Stderr = oldErr
	outW.Close()
	errW.Close()

	var outBuf, errBuf strings.Builder
	if _, err := io.Copy(&outBuf, outR); err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if _, err := io.Copy(&errBuf, errR); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	return outBuf.String(), errBuf.String()
}

func newTestLexer() *lexer {
	return &lexer{options: map[string]string{}}
}

func parseLine(t *testing.T, lx *lexer, line string) (string, string) {
	t.Helper()
	var out, errStr string
	out, errStr = runCaptured(t, func() {
		runShellLine(lx, line)
	})
	return out, errStr
}

func TestUnknownCommand(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "blah")
	if !strings.Contains(errStr, "unknown command 'blah'") {
		t.Errorf("want unknown command message, got: %q", errStr)
	}
}

func TestLooksLikeWallpaperID(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "94x38z")
	if !strings.Contains(errStr, "info 94x38z") {
		t.Errorf("want wallpaper-ID hint, got: %q", errStr)
	}
}

func TestMissingID(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "info")
	if !strings.Contains(errStr, "'info' takes exactly one wallpaper ID") {
		t.Errorf("want missing-ID message, got: %q", errStr)
	}
}

func TestExtraID(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "info 12345 67890")
	if !strings.Contains(errStr, "'info' takes exactly one wallpaper ID") {
		t.Errorf("want extra-ID message, got: %q", errStr)
	}
}

func TestEnvNoArg(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "env")
	if !strings.Contains(errStr, "'env' takes exactly one variable") {
		t.Errorf("want env message, got: %q", errStr)
	}
}

func TestEnvExtraArg(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "env query 123")
	if !strings.Contains(errStr, "'env' takes exactly one variable") {
		t.Errorf("want env message, got: %q", errStr)
	}
}

func TestMissingEquals(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "resolution 1920x1080")
	if !strings.Contains(errStr, "expected '=' after resolution") {
		t.Errorf("want missing-'=' message, got: %q", errStr)
	}
}

func TestUnknownVariable(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "foo = bar")
	if !strings.Contains(errStr, "unknown variable 'foo'") {
		t.Errorf("want unknown-variable message, got: %q", errStr)
	}
}

func TestValueAfterCommand(t *testing.T) {
	_, errStr := parseLine(t, newTestLexer(), "search extra")
	if !strings.Contains(errStr, "'search' takes no arguments") {
		t.Errorf("want search message, got: %q", errStr)
	}
}

func TestValidAssignment(t *testing.T) {
	lx := newTestLexer()
	out, errStr := parseLine(t, lx, "purity = 110")
	if errStr != "" {
		t.Errorf("want no error, got: %q", errStr)
	}
	if !strings.Contains(out, "purity is now [ 110 ]") {
		t.Errorf("want confirmation, got: %q", out)
	}
	if lx.options["purity"] != "110" {
		t.Errorf("option not stored: %q", lx.options["purity"])
	}
}

func TestCompactAssignment(t *testing.T) {
	lx := newTestLexer()
	out, errStr := parseLine(t, lx, "resolution=2560x1440")
	if errStr != "" {
		t.Errorf("want no error, got: %q", errStr)
	}
	if !strings.Contains(out, "resolution is now [ 2560x1440 ]") {
		t.Errorf("want confirmation, got: %q", out)
	}
	if lx.options["resolution"] != "2560x1440" {
		t.Errorf("option not stored: %q", lx.options["resolution"])
	}
}

func TestMultiWordValue(t *testing.T) {
	lx := newTestLexer()
	out, errStr := parseLine(t, lx, "query = +anime +summer -people")
	if errStr != "" {
		t.Errorf("want no error, got: %q", errStr)
	}
	if lx.options["query"] != "+anime +summer -people" {
		t.Errorf("multi-word value not stored correctly: %q", lx.options["query"])
	}
	if !strings.Contains(out, "query is now [ +anime +summer -people ]") {
		t.Errorf("want confirmation, got: %q", out)
	}
}

func TestInvalidValues(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{"purity = 99", "invalid value 99 for purity"},
		{"kind = abc", "invalid value abc for kind"},
		{"resolution = 1920", "invalid value 1920 for resolution"},
		{"color = pink", "invalid color value(s): pink"},
		{"sorting = newest", "invalid value newest for sorting"},
		{"order = up", "invalid value up for order"},
	}
	for _, c := range cases {
		lx := newTestLexer()
		_, errStr := parseLine(t, lx, c.line)
		if !strings.Contains(errStr, c.want) {
			t.Errorf("%s: want %q in error, got: %q", c.line, c.want, errStr)
		}
	}
}

func TestValidColors(t *testing.T) {
	for _, v := range []string{"red", "red,orange", "663399", "red,0066cc", ""} {
		ok, msg := validateValue("color", v)
		if !ok {
			t.Errorf("color %q should be valid: %s", v, msg)
		}
	}
}

func TestEnvShowsValue(t *testing.T) {
	lx := newTestLexer()
	lx.options["query"] = "anime"
	out, _ := parseLine(t, lx, "env query")
	if !strings.Contains(out, "Variable query is set to [ anime ]") {
		t.Errorf("want env output, got: %q", out)
	}
}

func TestEnvUnsetVariable(t *testing.T) {
	lx := newTestLexer()
	lx.options["resolution"] = "1920x1080"
	_, errStr := parseLine(t, lx, "env query")
	if !strings.Contains(errStr, "variable query is not set") {
		t.Errorf("want unset-variable message, got: %q", errStr)
	}
}

func TestHelp(t *testing.T) {
	out, errStr := parseLine(t, newTestLexer(), "help")
	if errStr != "" {
		t.Errorf("want no error, got: %q", errStr)
	}
	for _, want := range []string{"Shell commands:", "Settable variables", "info <ID>", "purity", "sorting"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing %q", want)
		}
	}
}

func TestConsecutiveParseErrors(t *testing.T) {
	lx := newTestLexer()
	// an error must not poison subsequent lines
	parseLine(t, lx, "blah")
	parseLine(t, lx, "info")
	out, errStr := parseLine(t, lx, "path = /tmp")
	if errStr != "" {
		t.Errorf("want no error after previous failures, got: %q", errStr)
	}
	if !strings.Contains(out, "path is now [ /tmp ]") {
		t.Errorf("want confirmation, got: %q", out)
	}
}

func TestClearValue(t *testing.T) {
	lx := newTestLexer()
	lx.options["query"] = "anime"
	out, errStr := parseLine(t, lx, "query =")
	if errStr != "" {
		t.Errorf("want no error, got: %q", errStr)
	}
	if !strings.Contains(out, "query has been cleared") {
		t.Errorf("want cleared message, got: %q", out)
	}
	if lx.options["query"] != "" {
		t.Errorf("query not cleared: %q", lx.options["query"])
	}

	// non-clearable variables reject the empty value
	out, errStr = parseLine(t, lx, "resolution =")
	if !strings.Contains(errStr, "value for resolution must not be empty") {
		t.Errorf("want empty-value rejection, got: %q", errStr)
	}
}

func TestApiErrorMessage(t *testing.T) {
	got := api.ApiErrorMessage("wallpaper", "zzz", &api.HTTPError{Status: 404})
	if !strings.Contains(got, "not found") {
		t.Errorf("want not-found message, got: %q", got)
	}
	got = api.ApiErrorMessage("search", "", &api.HTTPError{Status: 429})
	if !strings.Contains(got, "429") {
		t.Errorf("want 429 message, got: %q", got)
	}
}
