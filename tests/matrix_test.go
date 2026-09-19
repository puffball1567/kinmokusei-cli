package cli

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestBooleanOptionMatrix(t *testing.T) {
	accepted := []struct {
		name string
		args []string
		want bool
	}{
		{"implicit long", []string{"--enabled"}, true},
		{"implicit short", []string{"-e"}, true},
		{"numeric true", []string{"--enabled=1"}, true},
		{"short true", []string{"--enabled=t"}, true},
		{"upper short true", []string{"--enabled=T"}, true},
		{"upper true", []string{"--enabled=TRUE"}, true},
		{"title true", []string{"--enabled=True"}, true},
		{"lower true", []string{"--enabled=true"}, true},
		{"numeric false", []string{"--enabled=0"}, false},
		{"short false", []string{"--enabled=f"}, false},
		{"upper short false", []string{"--enabled=F"}, false},
		{"upper false", []string{"--enabled=FALSE"}, false},
		{"title false", []string{"--enabled=False"}, false},
		{"lower false", []string{"--enabled=false"}, false},
	}
	for _, test := range accepted {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag("enabled", "e", ""))
			parsed, err := command.Parse(test.args)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if got := parsed.Flag("enabled"); got != test.want || !parsed.Has("enabled") {
				t.Errorf("enabled = (%v, provided %v), want (%v, true)", got, parsed.Has("enabled"), test.want)
			}
		})
	}

	for _, value := range []string{"", "yes", "no", "2", " true "} {
		t.Run("reject "+strconv.Quote(value), func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag("enabled", "e", ""))
			_, err := command.Parse([]string{"--enabled=" + value})
			if err == nil || !strings.Contains(err.Error(), "expected boolean") {
				t.Fatalf("Parse() error = %v, want expected boolean", err)
			}
		})
	}
}

func TestTextOptionMatrix(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"long separate", []string{"--output", "file.txt"}, "file.txt"},
		{"long inline", []string{"--output=file.txt"}, "file.txt"},
		{"short separate", []string{"-o", "file.txt"}, "file.txt"},
		{"empty inline", []string{"--output="}, ""},
		{"option-looking separate", []string{"--output", "--help"}, "--help"},
		{"option-looking inline", []string{"--output=-value"}, "-value"},
		{"unicode", []string{"--output", "結果.txt"}, "結果.txt"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(RequiredText("output", "o", ""))
			parsed, err := command.Parse(test.args)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if got := parsed.Text("output"); got != test.want || !parsed.Has("output") {
				t.Errorf("output = (%q, provided %v), want (%q, true)", got, parsed.Has("output"), test.want)
			}
			if parsed.HelpRequested {
				t.Errorf("value %q was interpreted as help", test.want)
			}
		})
	}
}

func TestIntegerOptionMatrix(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	accepted := []struct {
		name  string
		value string
		want  int
	}{
		{"zero", "0", 0},
		{"positive", "42", 42},
		{"explicit positive", "+42", 42},
		{"negative", "-42", -42},
		{"maximum", strconv.Itoa(maxInt), maxInt},
		{"minimum", strconv.Itoa(minInt), minInt},
	}
	for _, test := range accepted {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(RequiredInteger("count", "c", ""))
			parsed, err := command.Parse([]string{"--count", test.value})
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.value, err)
			}
			if got := parsed.Integer("count"); got != test.want || !parsed.Has("count") {
				t.Errorf("count = (%d, provided %v), want (%d, true)", got, parsed.Has("count"), test.want)
			}
		})
	}

	tooLarge := strconv.FormatUint(uint64(maxInt)+1, 10)
	tooSmall := "-" + strconv.FormatUint(uint64(maxInt)+2, 10)
	for _, value := range []string{"", "1.5", "0x10", " 1", "1 ", tooLarge, tooSmall} {
		t.Run("reject "+strconv.Quote(value), func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(RequiredInteger("count", "c", ""))
			_, err := command.Parse([]string{"--count=" + value})
			if err == nil || !strings.Contains(err.Error(), "expected integer") {
				t.Fatalf("Parse() error = %v, want expected integer", err)
			}
		})
	}
}

func TestPositionAndOptionOrderingMatrix(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantFlag    bool
		positionals []string
	}{
		{"empty", nil, false, []string{}},
		{"only positionals", []string{"a", "b"}, false, []string{"a", "b"}},
		{"option before positional", []string{"--verbose", "a"}, true, []string{"a"}},
		{"option after positional", []string{"a", "--verbose", "b"}, true, []string{"a", "b"}},
		{"end marker", []string{"--", "--verbose"}, false, []string{"--verbose"}},
		{"repeated end marker", []string{"--", "--", "-v"}, false, []string{"--", "-v"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag("verbose", "v", ""))
			parsed, err := command.Parse(test.args)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if parsed.Flag("verbose") != test.wantFlag {
				t.Errorf("verbose = %v, want %v", parsed.Flag("verbose"), test.wantFlag)
			}
			if !reflect.DeepEqual(parsed.Positionals, test.positionals) {
				t.Errorf("Positionals = %#v, want %#v", parsed.Positionals, test.positionals)
			}
		})
	}
}

func TestHelpAndVersionControlMatrix(t *testing.T) {
	command := NewCommand("tool", "")
	command.SetVersion("1.0.0")
	command.Option(RequiredText("output", "o", ""))

	tests := []struct {
		name        string
		args        []string
		help        bool
		version     bool
		wantErrPart string
	}{
		{"long help", []string{"--help"}, true, false, ""},
		{"short help", []string{"-h"}, true, false, ""},
		{"version", []string{"--version"}, false, true, ""},
		{"both", []string{"--help", "--version"}, true, true, ""},
		{"repeated help", []string{"--help", "-h"}, true, false, ""},
		{"help does not hide unknown", []string{"--help", "--missing"}, false, false, "unknown option"},
		{"version does not hide unknown", []string{"--version", "--missing"}, false, false, "unknown option"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := command.Parse(test.args)
			if test.wantErrPart != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErrPart) {
					t.Fatalf("Parse() error = %v, want %q", err, test.wantErrPart)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if parsed.HelpRequested != test.help || parsed.VersionRequested != test.version {
				t.Errorf("control flags = (%v, %v), want (%v, %v)", parsed.HelpRequested, parsed.VersionRequested, test.help, test.version)
			}
		})
	}
}

func TestUnknownGettersReturnZeroValues(t *testing.T) {
	parsed, err := NewCommand("tool", "").Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Has("missing") || parsed.Text("missing") != "" || parsed.Integer("missing") != 0 || parsed.Flag("missing") {
		t.Fatalf("unknown getters did not return zero values")
	}
}

func TestSubcommandDispatchMatrix(t *testing.T) {
	root := NewCommand("tool", "")
	root.Option(Flag("verbose", "v", ""))
	serve := NewCommand("serve", "")
	serve.Option(RequiredInteger("port", "p", ""))
	root.Subcommand(serve)

	t.Run("first argument dispatches", func(t *testing.T) {
		parsed, err := root.Parse([]string{"serve", "--port", "8080"})
		if err != nil || parsed.Command != "serve" || parsed.Integer("port") != 8080 {
			t.Fatalf("Parse() = (%#v, %v)", parsed, err)
		}
	})
	t.Run("child help bypasses child required", func(t *testing.T) {
		parsed, err := root.Parse([]string{"serve", "--help"})
		if err != nil || parsed.Command != "serve" || !parsed.HelpRequested {
			t.Fatalf("Parse() = (%#v, %v)", parsed, err)
		}
	})
	t.Run("child parse error propagates", func(t *testing.T) {
		_, err := root.Parse([]string{"serve"})
		if err == nil || !strings.Contains(err.Error(), "required option --port is missing") {
			t.Fatalf("Parse() error = %v", err)
		}
	})
	t.Run("root option keeps later command positional", func(t *testing.T) {
		parsed, err := root.Parse([]string{"--verbose", "serve"})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if parsed.Command != "tool" || !parsed.Flag("verbose") || !reflect.DeepEqual(parsed.Positionals, []string{"serve"}) {
			t.Fatalf("Parse() = %#v", parsed)
		}
	})
	t.Run("option-like unknown stays root option error", func(t *testing.T) {
		_, err := root.Parse([]string{"--missing", "serve"})
		if err == nil || !strings.Contains(err.Error(), "unknown option") {
			t.Fatalf("Parse() error = %v", err)
		}
	})
}

func TestChildDefinitionErrorPropagates(t *testing.T) {
	root := NewCommand("tool", "")
	child := NewCommand("serve", "")
	child.Option(Flag("bad name", "", ""))
	root.Subcommand(child)
	_, err := root.Parse(nil)
	if err == nil || !strings.Contains(err.Error(), "invalid option name") {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestOptionNameValidationMatrix(t *testing.T) {
	for _, name := range []string{"name", "dry-run", "UPPER", "with.dot", "with_underscore", "n1"} {
		t.Run("accept "+name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag(name, "", ""))
			if _, err := command.Parse(nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
	for _, name := range []string{"", " ", " leading", "trailing ", "internal space", "tab\tname", "line\nname", "-name", "name=value", "name/path", "name\\path"} {
		t.Run("reject "+strconv.Quote(name), func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag(name, "", ""))
			if _, err := command.Parse(nil); err == nil || !strings.Contains(err.Error(), "invalid option name") {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestShortNameValidationMatrix(t *testing.T) {
	for _, short := range []string{"", "a", "Z", "7"} {
		t.Run("accept "+strconv.Quote(short), func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag("option", short, ""))
			if _, err := command.Parse(nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
	for _, short := range []string{" ", "-", "ab", "あ"} {
		t.Run("reject "+strconv.Quote(short), func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(Flag("option", short, ""))
			if _, err := command.Parse(nil); err == nil || !strings.Contains(err.Error(), "invalid short option") {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestCommandNameValidationMatrix(t *testing.T) {
	for _, name := range []string{"tool", "dry-run", "UPPER", "tool.v2", "tool_name"} {
		t.Run("accept "+name, func(t *testing.T) {
			if _, err := NewCommand(name, "").Parse(nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
	for _, name := range []string{"", " ", " leading", "trailing ", "internal space", "tab\tname", "line\nname", "-name", "name/path", "name\\path"} {
		t.Run("reject "+strconv.Quote(name), func(t *testing.T) {
			if _, err := NewCommand(name, "").Parse(nil); err == nil {
				t.Fatalf("Parse() error = nil")
			}
		})
	}
}

func TestVersionNameAlwaysReserved(t *testing.T) {
	for _, configuredVersion := range []string{"", "1.0.0"} {
		t.Run(strconv.Quote(configuredVersion), func(t *testing.T) {
			command := NewCommand("tool", "")
			if configuredVersion != "" {
				command.SetVersion(configuredVersion)
			}
			command.Option(Flag("version", "", ""))
			_, err := command.Parse(nil)
			if err == nil || !strings.Contains(err.Error(), "duplicate option --version") {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

func TestMinimalHelp(t *testing.T) {
	want := "Usage: tool [arguments]\n\nOptions:\n  -h, --help\tShow help\n"
	if got := NewCommand("tool", "").Help(); got != want {
		t.Fatalf("Help() = %q, want %q", got, want)
	}
}

func TestHelpForLongOnlyOption(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(Text("output", "", "", "stdout"))
	want := "Usage: tool [options] [arguments]\n\nOptions:\n" +
		"  -h, --help\tShow help\n" +
		"      --output <text>\n"
	if got := command.Help(); got != want {
		t.Fatalf("Help() = %q, want %q", got, want)
	}
}

func FuzzParseNeverPanics(f *testing.F) {
	for _, seed := range []string{
		"",
		"--help",
		"--count\x00-1\x00--output=結果.txt",
		"--include\x00src\x00--include=tests\x00--port\x0080\x00-p\x00443",
		"--\x00--count\x001",
		"-vc",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, encoded string) {
		command := NewCommand("tool", "")
		command.SetVersion("1.0.0")
		command.Option(Flag("verbose", "v", ""))
		command.Option(Integer("count", "c", "", 0))
		command.Option(Text("output", "o", "", ""))
		command.Option(TextList("include", "I", ""))
		command.Option(IntegerList("port", "p", ""))
		command.Option(Choice("format", "f", "", []string{"text", "json"}, "text"))
		command.Option(IntegerRange("jobs", "j", "", 2, 1, 4))
		_, _ = command.Parse(strings.Split(encoded, "\x00"))
	})
}

func FuzzDefinitionValidationNeverPanics(f *testing.F) {
	for _, seed := range []struct {
		command string
		option  string
		short   string
	}{
		{"tool", "output", "o"},
		{"", "", ""},
		{"-tool", "bad name", "あ"},
	} {
		f.Add(seed.command, seed.option, seed.short)
	}
	f.Fuzz(func(t *testing.T, commandName string, optionName string, short string) {
		command := NewCommand(commandName, "")
		command.Option(Flag(optionName, short, ""))
		command.SetArgumentRange(len(optionName)%4-1, len(short)%5-2)
		_, _ = command.Parse(nil)
	})
}
