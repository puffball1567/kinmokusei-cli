package cli

import (
	"reflect"
	"strings"
	"testing"
)

func TestRepeatableOptionMatrix(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantIncludes []string
		wantPorts    []int
	}{
		{"none", nil, []string{}, []int{}},
		{"one each", []string{"--include", "src", "--port=80"}, []string{"src"}, []int{80}},
		{"mixed forms preserve order", []string{"-I", "src", "--include=tests", "-p", "80", "--port", "443"}, []string{"src", "tests"}, []int{80, 443}},
		{"empty and unicode text", []string{"--include=", "--include", "資料"}, []string{"", "資料"}, []int{}},
		{"signed integers", []string{"--port", "-1", "--port=+2", "-p", "0"}, []string{}, []int{-1, 2, 0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.Option(TextList("include", "I", "Include path"))
			command.Option(IntegerList("port", "p", "Port"))
			parsed, err := command.Parse(test.args)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.args, err)
			}
			if got := parsed.Texts("include"); !reflect.DeepEqual(got, test.wantIncludes) {
				t.Errorf("Texts(include) = %#v, want %#v", got, test.wantIncludes)
			}
			if got := parsed.Integers("port"); !reflect.DeepEqual(got, test.wantPorts) {
				t.Errorf("Integers(port) = %#v, want %#v", got, test.wantPorts)
			}
			if parsed.Has("include") != (len(test.wantIncludes) > 0) {
				t.Errorf("Has(include) = %v", parsed.Has("include"))
			}
			if parsed.Has("port") != (len(test.wantPorts) > 0) {
				t.Errorf("Has(port) = %v", parsed.Has("port"))
			}
		})
	}
}

func TestRequiredRepeatableOptions(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(RequiredTextList("include", "I", ""))
	command.Option(RequiredIntegerList("port", "p", ""))

	if _, err := command.Parse(nil); err == nil || !strings.Contains(err.Error(), "required option --include is missing") {
		t.Fatalf("Parse(nil) error = %v", err)
	}
	if _, err := command.Parse([]string{"--include", "src"}); err == nil || !strings.Contains(err.Error(), "required option --port is missing") {
		t.Fatalf("Parse(include) error = %v", err)
	}
	parsed, err := command.Parse([]string{"--include", "src", "--port", "8080"})
	if err != nil || !parsed.Has("include") || !parsed.Has("port") {
		t.Fatalf("Parse(required lists) = (%#v, %v)", parsed, err)
	}
}

func TestRepeatableIntegerRejectsInvalidValues(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(IntegerList("port", "p", ""))
	_, err := command.Parse([]string{"--port", "80", "--port", "many"})
	if err == nil || !strings.Contains(err.Error(), `invalid value "many" for --port: expected integer`) {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestUnknownListGettersReturnEmptyValues(t *testing.T) {
	parsed, err := NewCommand("tool", "").Parse(nil)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Texts("missing") != nil || parsed.Integers("missing") != nil {
		t.Fatalf("unknown list getters = (%#v, %#v), want nil slices", parsed.Texts("missing"), parsed.Integers("missing"))
	}
}

func TestArgumentRangeMatrix(t *testing.T) {
	tests := []struct {
		name    string
		minimum int
		maximum int
		args    []string
		wantErr string
	}{
		{"exact zero accepts zero", 0, 0, nil, ""},
		{"exact zero rejects one", 0, 0, []string{"a"}, "expected at most 0 positional argument(s), got 1"},
		{"unbounded accepts minimum", 1, -1, []string{"a"}, ""},
		{"unbounded accepts many", 1, -1, []string{"a", "b", "c"}, ""},
		{"unbounded rejects below minimum", 1, -1, nil, "expected at least 1 positional argument(s), got 0"},
		{"bounded accepts lower boundary", 2, 3, []string{"a", "b"}, ""},
		{"bounded accepts upper boundary", 2, 3, []string{"a", "b", "c"}, ""},
		{"bounded rejects below", 2, 3, []string{"a"}, "expected at least 2 positional argument(s), got 1"},
		{"bounded rejects above", 2, 3, []string{"a", "b", "c", "d"}, "expected at most 3 positional argument(s), got 4"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.SetArgumentRange(test.minimum, test.maximum)
			parsed, err := command.Parse(test.args)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("Parse(%q) error = %v, want %q", test.args, err, test.wantErr)
				}
				return
			}
			wantPositionals := append([]string{}, test.args...)
			if err != nil || !reflect.DeepEqual(parsed.Positionals, wantPositionals) {
				t.Fatalf("Parse(%q) = (%#v, %v)", test.args, parsed, err)
			}
		})
	}
}

func TestArgumentRangeDefinitionErrors(t *testing.T) {
	tests := []struct {
		name    string
		minimum int
		maximum int
		want    string
	}{
		{"negative minimum", -1, -1, "minimum argument count must not be negative"},
		{"maximum below unbounded sentinel", 0, -2, "maximum argument count must be -1 or greater"},
		{"maximum below minimum", 2, 1, "maximum argument count must not be less than minimum"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := NewCommand("tool", "")
			command.SetArgumentRange(test.minimum, test.maximum)
			_, err := command.Parse(nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestControlRequestsBypassArgumentRange(t *testing.T) {
	command := NewCommand("tool", "")
	command.SetVersion("1.0.0")
	command.SetArgumentRange(2, 2)
	for _, args := range [][]string{{"--help"}, {"--version"}} {
		if _, err := command.Parse(args); err != nil {
			t.Fatalf("Parse(%q) error = %v", args, err)
		}
	}
}

func TestCommandPathMatrix(t *testing.T) {
	root := NewCommand("tool", "")
	config := NewCommand("config", "")
	set := NewCommand("set", "")
	config.Subcommand(set)
	root.Subcommand(config)

	tests := []struct {
		name string
		cmd  *Command
		args []string
		want []string
	}{
		{"root", root, nil, []string{"tool"}},
		{"child", root, []string{"config"}, []string{"tool", "config"}},
		{"nested", root, []string{"config", "set"}, []string{"tool", "config", "set"}},
		{"direct child object", config, []string{"set"}, []string{"config", "set"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, err := test.cmd.Parse(test.args)
			if err != nil || !reflect.DeepEqual(parsed.CommandPath, test.want) {
				t.Fatalf("Parse(%q) path = (%#v, %v), want %#v", test.args, parsed.CommandPath, err, test.want)
			}
		})
	}
}

func TestAdvancedHelpFormatting(t *testing.T) {
	command := NewCommand("tool", "")
	command.SetArgumentRange(1, 3)
	command.Option(TextList("include", "I", "Include path"))
	command.Option(IntegerList("port", "", "Port"))
	want := "Usage: tool [options] <arguments>\n" +
		"\nOptions:\n" +
		"  -h, --help\tShow help\n" +
		"  -I, --include <text>...\tInclude path\n" +
		"      --port <integer>...\tPort\n"
	if got := command.Help(); got != want {
		t.Fatalf("Help() =\n%q\nwant\n%q", got, want)
	}

	noArguments := NewCommand("tool", "")
	noArguments.SetArgumentRange(0, 0)
	if got, want := noArguments.Help(), "Usage: tool\n\nOptions:\n  -h, --help\tShow help\n"; got != want {
		t.Fatalf("Help() = %q, want %q", got, want)
	}
}
