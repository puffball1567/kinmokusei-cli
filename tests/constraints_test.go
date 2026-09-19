package cli

import (
	"strings"
	"testing"
)

func TestChoiceOptionMatrix(t *testing.T) {
	choices := []string{"text", "json", "日本語", ""}
	tests := []struct {
		name    string
		args    []string
		want    string
		has     bool
		wantErr string
	}{
		{"default", nil, "text", false, ""},
		{"long separate", []string{"--format", "json"}, "json", true, ""},
		{"long inline", []string{"--format=日本語"}, "日本語", true, ""},
		{"short", []string{"-f", "json"}, "json", true, ""},
		{"empty allowed", []string{"--format="}, "", true, ""},
		{"case sensitive rejection", []string{"--format", "JSON"}, "", false, "expected one of text, json, 日本語, "},
		{"unknown rejection", []string{"--format", "xml"}, "", false, "expected one of text, json, 日本語, "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Choice("format", "f", "Output format", choices, "text"))
			parsed, err := command.Parse(test.args)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("Parse(%q) error = %v, want %q", test.args, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if got := parsed.Text("format"); got != test.want || parsed.Has("format") != test.has {
				t.Errorf("format = (%q, provided %v), want (%q, %v)", got, parsed.Has("format"), test.want, test.has)
			}
		})
	}
}

func TestRequiredChoice(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(RequiredChoice("color", "c", "", []string{"auto", "always", "never"}))
	if _, err := command.Parse(nil); err == nil || !strings.Contains(err.Error(), "required option --color is missing") {
		t.Fatalf("Parse(nil) error = %v", err)
	}
	parsed, err := command.Parse([]string{"--color", "never"})
	if err != nil || parsed.Text("color") != "never" || !parsed.Has("color") {
		t.Fatalf("Parse(color) = (%#v, %v)", parsed, err)
	}
}

func TestChoiceDefinitionValidation(t *testing.T) {
	tests := []struct {
		name  string
		build func() Option
		want  string
	}{
		{"empty choices", func() Option { return RequiredChoice("format", "f", "", nil) }, "must define at least one choice"},
		{"duplicate choices", func() Option { return Choice("format", "f", "", []string{"json", "text", "json"}, "json") }, `duplicate choice "json"`},
		{"duplicate empty choices", func() Option { return Choice("format", "f", "", []string{"", ""}, "") }, `duplicate choice ""`},
		{"invalid default", func() Option { return Choice("format", "f", "", []string{"json", "text"}, "xml") }, `default value "xml"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(test.build())
			_, err := command.Parse(nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestIntegerRangeOptionMatrix(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    int
		has     bool
		wantErr string
	}{
		{"default", nil, 4, false, ""},
		{"minimum", []string{"--jobs", "1"}, 1, true, ""},
		{"maximum", []string{"--jobs=8"}, 8, true, ""},
		{"middle short", []string{"-j", "5"}, 5, true, ""},
		{"below minimum", []string{"--jobs", "0"}, 0, false, "expected 1..8"},
		{"above maximum", []string{"--jobs", "9"}, 0, false, "expected 1..8"},
		{"not an integer", []string{"--jobs", "many"}, 0, false, "expected integer"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(IntegerRange("jobs", "j", "Parallel jobs", 4, 1, 8))
			parsed, err := command.Parse(test.args)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("Parse(%q) error = %v, want %q", test.args, err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if got := parsed.Integer("jobs"); got != test.want || parsed.Has("jobs") != test.has {
				t.Errorf("jobs = (%d, provided %v), want (%d, %v)", got, parsed.Has("jobs"), test.want, test.has)
			}
		})
	}
}

func TestRequiredIntegerRange(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(RequiredIntegerRange("level", "l", "", -2, 2))
	if _, err := command.Parse(nil); err == nil || !strings.Contains(err.Error(), "required option --level is missing") {
		t.Fatalf("Parse(nil) error = %v", err)
	}
	parsed, err := command.Parse([]string{"--level", "-2"})
	if err != nil || parsed.Integer("level") != -2 || !parsed.Has("level") {
		t.Fatalf("Parse(level) = (%#v, %v)", parsed, err)
	}
}

func TestIntegerRangeDefinitionValidation(t *testing.T) {
	tests := []struct {
		name string
		spec Option
		want string
	}{
		{"reversed range", RequiredIntegerRange("jobs", "j", "", 8, 1), "maximum value for --jobs must not be less than minimum"},
		{"default below", IntegerRange("jobs", "j", "", 0, 1, 8), "default value 0 for --jobs is outside 1..8"},
		{"default above", IntegerRange("jobs", "j", "", 9, 1, 8), "default value 9 for --jobs is outside 1..8"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(test.spec)
			_, err := command.Parse(nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want %q", err, test.want)
			}
		})
	}

	exact := NewCommand("tool", "")
	exact.Option(IntegerRange("level", "l", "", 3, 3, 3))
	if parsed, err := exact.Parse(nil); err != nil || parsed.Integer("level") != 3 {
		t.Fatalf("exact range Parse() = (%#v, %v)", parsed, err)
	}
}

func TestConstrainedOptionHelp(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(Choice("format", "f", "Output format", []string{"text", "json"}, "text"))
	command.Option(RequiredIntegerRange("jobs", "", "Parallel jobs", 1, 8))
	want := "Usage: tool [options] [arguments]\n" +
		"\nOptions:\n" +
		"  -h, --help\tShow help\n" +
		"  -f, --format <text|json>\tOutput format\n" +
		"      --jobs <integer:1..8>\tParallel jobs (required)\n"
	if got := command.Help(); got != want {
		t.Fatalf("Help() =\n%q\nwant\n%q", got, want)
	}
}
