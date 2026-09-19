package cli

import (
	"reflect"
	"strings"
	"testing"
)

func newTestCommand() *Command {
	command := NewCommand("tool", "A test command")
	command.SetVersion("1.2.3")
	command.Option(Flag("verbose", "v", "Enable verbose output"))
	command.Option(Text("format", "f", "Output format", "text"))
	command.Option(Integer("count", "c", "Repeat count", 1))
	command.Option(RequiredText("output", "o", "Output path"))
	return command
}

func TestParseAllOptionKindsAndPositionals(t *testing.T) {
	parsed, err := newTestCommand().Parse([]string{
		"--verbose", "--format=json", "-c", "-2", "--output", "report.txt", "input-a", "入力-b",
	})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Command != "tool" {
		t.Errorf("Command = %q, want tool", parsed.Command)
	}
	if !parsed.Flag("verbose") || !parsed.Has("verbose") {
		t.Errorf("verbose = (%v, provided %v), want (true, true)", parsed.Flag("verbose"), parsed.Has("verbose"))
	}
	if got := parsed.Text("format"); got != "json" {
		t.Errorf("format = %q, want json", got)
	}
	if got := parsed.Integer("count"); got != -2 {
		t.Errorf("count = %d, want -2", got)
	}
	if got := parsed.Text("output"); got != "report.txt" {
		t.Errorf("output = %q, want report.txt", got)
	}
	if want := []string{"input-a", "入力-b"}; !reflect.DeepEqual(parsed.Positionals, want) {
		t.Errorf("Positionals = %#v, want %#v", parsed.Positionals, want)
	}
}

func TestDefaultsAndExplicitFalse(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(Flag("color", "", ""))
	command.Option(Text("format", "", "", "text"))
	command.Option(Integer("count", "", "", 3))

	parsed, err := command.Parse([]string{"--color=false", "-", "tail"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Flag("color") || !parsed.Has("color") {
		t.Errorf("color = (%v, provided %v), want (false, true)", parsed.Flag("color"), parsed.Has("color"))
	}
	if got := parsed.Text("format"); got != "text" || parsed.Has("format") {
		t.Errorf("format = (%q, provided %v), want (text, false)", got, parsed.Has("format"))
	}
	if got := parsed.Integer("count"); got != 3 || parsed.Has("count") {
		t.Errorf("count = (%d, provided %v), want (3, false)", got, parsed.Has("count"))
	}
	if want := []string{"-", "tail"}; !reflect.DeepEqual(parsed.Positionals, want) {
		t.Errorf("Positionals = %#v, want %#v", parsed.Positionals, want)
	}
}

func TestEndOfOptionsPreservesOptionLikeArguments(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(Flag("verbose", "v", ""))
	parsed, err := command.Parse([]string{"--", "--verbose", "-v", ""})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Flag("verbose") || parsed.Has("verbose") {
		t.Errorf("verbose was parsed after --")
	}
	if want := []string{"--verbose", "-v", ""}; !reflect.DeepEqual(parsed.Positionals, want) {
		t.Errorf("Positionals = %#v, want %#v", parsed.Positionals, want)
	}
}

func TestHelpAndVersionBypassRequiredOptions(t *testing.T) {
	command := newTestCommand()
	help, err := command.Parse([]string{"--help"})
	if err != nil || !help.HelpRequested || help.VersionRequested {
		t.Fatalf("help Parse() = (%#v, %v)", help, err)
	}
	version, err := command.Parse([]string{"--version"})
	if err != nil || version.HelpRequested || !version.VersionRequested {
		t.Fatalf("version Parse() = (%#v, %v)", version, err)
	}
}

func TestSubcommand(t *testing.T) {
	root := NewCommand("tool", "")
	serve := NewCommand("serve", "Serve requests")
	serve.Option(RequiredInteger("port", "p", "Port"))
	root.Subcommand(serve)

	parsed, err := root.Parse([]string{"serve", "--port", "8080", "public"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Command != "serve" || parsed.Integer("port") != 8080 {
		t.Errorf("parsed = (%q, %d), want (serve, 8080)", parsed.Command, parsed.Integer("port"))
	}
	if want := []string{"public"}; !reflect.DeepEqual(parsed.Positionals, want) {
		t.Errorf("Positionals = %#v, want %#v", parsed.Positionals, want)
	}
}

func TestNestedSubcommandKeepsLeafCommand(t *testing.T) {
	root := NewCommand("tool", "")
	config := NewCommand("config", "")
	set := NewCommand("set", "")
	set.Option(RequiredText("key", "k", ""))
	config.Subcommand(set)
	root.Subcommand(config)

	parsed, err := root.Parse([]string{"config", "set", "--key", "theme"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.Command != "set" || parsed.Text("key") != "theme" {
		t.Errorf("parsed = (%q, %q), want (set, theme)", parsed.Command, parsed.Text("key"))
	}
}

func TestEmptyInlineValueSatisfiesRequiredText(t *testing.T) {
	command := NewCommand("tool", "")
	command.Option(RequiredText("output", "o", ""))
	parsed, err := command.Parse([]string{"--output="})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !parsed.Has("output") || parsed.Text("output") != "" {
		t.Errorf("output = (%q, provided %v), want empty explicit value", parsed.Text("output"), parsed.Has("output"))
	}
}

func TestParseFailures(t *testing.T) {
	tests := []struct {
		name    string
		command func() *Command
		args    []string
		want    string
	}{
		{"unknown long", newTestCommand, []string{"--missing"}, "unknown option --missing"},
		{"unknown short", newTestCommand, []string{"-x"}, "unknown option -x"},
		{"grouped short", newTestCommand, []string{"-vc"}, "short options cannot be grouped"},
		{"missing value", newTestCommand, []string{"--output"}, "option --output requires a value"},
		{"invalid integer", newTestCommand, []string{"--count", "many", "--output", "x"}, "invalid value \"many\" for --count: expected integer"},
		{"invalid boolean", newTestCommand, []string{"--verbose=perhaps", "--output", "x"}, "invalid value \"perhaps\" for --verbose: expected boolean"},
		{"duplicate use", newTestCommand, []string{"--output", "a", "--output", "b"}, "may only be specified once"},
		{"required missing", newTestCommand, nil, "required option --output is missing"},
		{"version disabled", func() *Command { return NewCommand("tool", "") }, []string{"--version"}, "unknown option --version"},
		{"unknown command", func() *Command {
			root := NewCommand("tool", "")
			root.Subcommand(NewCommand("serve", ""))
			return root
		}, []string{"missing"}, "unknown command \"missing\""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.command().Parse(test.args)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestDefinitionValidation(t *testing.T) {
	tests := []struct {
		name  string
		build func() *Command
		want  string
	}{
		{"empty command", func() *Command { return NewCommand(" ", "") }, "command name must not be empty"},
		{"empty option", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("", "", ""))
			return command
		}, "invalid option name"},
		{"hyphen option", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("--bad", "", ""))
			return command
		}, "invalid option name"},
		{"equals option", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("bad=name", "", ""))
			return command
		}, "invalid option name"},
		{"spaced option", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("bad name", "", ""))
			return command
		}, "invalid option name"},
		{"path option", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("bad/name", "", ""))
			return command
		}, "invalid option name"},
		{"duplicate long", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("verbose", "v", ""))
			command.Option(Text("verbose", "f", "", ""))
			return command
		}, "duplicate option --verbose"},
		{"duplicate short", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("verbose", "v", ""))
			command.Option(Text("value", "v", "", ""))
			return command
		}, "duplicate short option -v"},
		{"reserved help", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("help", "", ""))
			return command
		}, "duplicate option --help"},
		{"reserved short help", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("hello", "h", ""))
			return command
		}, "duplicate short option -h"},
		{"long short name", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("verbose", "vv", ""))
			return command
		}, "invalid short option"},
		{"space short name", func() *Command {
			command := NewCommand("tool", "")
			command.Option(Flag("verbose", " ", ""))
			return command
		}, "invalid short option"},
		{"reserved version", func() *Command {
			command := NewCommand("tool", "")
			command.SetVersion("1.0.0")
			command.Option(Flag("version", "", ""))
			return command
		}, "duplicate option --version"},
		{"duplicate command", func() *Command {
			command := NewCommand("tool", "")
			command.Subcommand(NewCommand("serve", ""))
			command.Subcommand(NewCommand("serve", ""))
			return command
		}, "duplicate command \"serve\""},
		{"invalid child command", func() *Command {
			command := NewCommand("tool", "")
			command.Subcommand(NewCommand("-serve", ""))
			return command
		}, "invalid command name"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := test.build().Parse(nil)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Parse() error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestHelp(t *testing.T) {
	command := newTestCommand()
	command.Subcommand(NewCommand("serve", "Serve requests"))
	want := "Usage: tool [options] <command> [arguments]\n" +
		"\nA test command\n" +
		"\nCommands:\n" +
		"  serve\tServe requests\n" +
		"\nOptions:\n" +
		"  -h, --help\tShow help\n" +
		"      --version\tShow version\n" +
		"  -v, --verbose\tEnable verbose output\n" +
		"  -f, --format <text>\tOutput format\n" +
		"  -c, --count <integer>\tRepeat count\n" +
		"  -o, --output <text>\tOutput path (required)\n"
	if got := command.Help(); got != want {
		t.Fatalf("Help() =\n%q\nwant\n%q", got, want)
	}
}
