# kinmokusei-cli

`kinmokusei-cli` is a command-line argument parser written for Kinmokusei. It
provides a Kinmokusei-first API while keeping process I/O and exit behavior
under application control.

The project is under active development.

## Requirements

- Kinmokusei v0.4.3 or later for automatic short imports
- Go 1.23 or later

The library itself remains compatible with Kinmokusei v0.4.1. Applications
using v0.4.1 or v0.4.2 must import the canonical module path instead.

## Features

- Long options such as `--output result.txt` and `--output=result.txt`
- Single-character short options such as `-o result.txt`
- Boolean, text, integer, enumerated-choice, and ranged-integer values
- Repeatable text and integer options
- Required options and typed defaults
- Positional arguments and the `--` end-of-options marker
- Minimum and maximum positional-argument counts
- Built-in help and version requests
- Nested command definitions
- Full command paths in parse results
- Definition checks for invalid and duplicate option or command names
- Parse errors returned through `Result<T>` rather than process termination

## Example

Add the tagged package to a Kinmokusei application and import its public module:

```sh
keika deps add github.com/puffball1567/kinmokusei-cli@v0.1.1
```

```ts
import { Command, Flag, IntegerRange, RequiredText } from "kinmokusei-cli";
import go fmt from "fmt";
import go os from "os";

function main(): void {
  const command = new Command("greeter", "Print a greeting");
  command.setVersion("0.1.0");
  command.option(Flag("loud", "l", "Use uppercase output"));
  command.option(IntegerRange("count", "c", "Number of greetings", 1, 1, 100));
  command.option(RequiredText("name", "n", "Name to greet"));

  const [parsed, err] = command.parse(os.Args[1:]);
  if (err !== nil) {
    fmt.Fprintln(os.Stderr, "error:", err.Error());
    return;
  }
  if (parsed.helpRequested) {
    fmt.Print(command.help());
    return;
  }

  for (let index = 0; index < parsed.integer("count"); index++) {
    fmt.Println("Hello, " + parsed.text("name") + "!");
  }
}
```

The parser receives application arguments without the executable name. It does
not read `os.Args`, write output, or call `os.Exit` itself.

See [`examples/basic/main.km`](examples/basic/main.km) for a complete example.
[`examples/subcommands/main.km`](examples/subcommands/main.km) demonstrates
nested dispatch, repeatable options, positional argument validation, and
command-specific help.

## Repeatable options

Use `TextList` or `IntegerList` when an option may appear more than once. The
values retain command-line order.

```ts
command.option(TextList("include", "I", "Include path"));
command.option(IntegerList("port", "p", "Port to publish"));

const [parsed, err] = command.parse([
  "--include", "src",
  "--include=tests",
  "-p", "8080",
]);
if (err !== nil) {
  return;
}

const includes = parsed.texts("include");
const ports = parsed.integers("port");
```

`RequiredTextList` and `RequiredIntegerList` require at least one occurrence.
Other option kinds continue to reject duplicate occurrences.

## Value constraints

`Choice` restricts a text option to an explicit list and validates its default.
`RequiredChoice` provides the same check without a default value.

```ts
command.option(Choice(
  "format",
  "f",
  "Output format",
  ["text", "json"],
  "text",
));
```

`IntegerRange` and `RequiredIntegerRange` accept inclusive lower and upper
bounds. Invalid ranges and defaults outside the range are rejected as command
definition errors, before arguments are parsed.

```ts
command.option(IntegerRange("jobs", "j", "Parallel jobs", 4, 1, 16));
```

## Positional arguments and command paths

`setArgumentRange(minimum, maximum)` validates the positional argument count.
Use `-1` as the maximum for no upper limit. Help and version requests bypass
this check in the same way that they bypass required options.

For subcommands, `command` is the selected leaf name and `commandPath` contains
the complete path. Parsing `tool config set` therefore returns `set` and
`["tool", "config", "set"]`, respectively.

## Parsing behavior

- `--help` and `-h` set `helpRequested`.
- `--version` sets `versionRequested` when a version is configured.
- Help and version requests do not require otherwise mandatory options.
- Scalar options may be supplied once. Repeated scalar options return an error;
  list options are explicitly repeatable.
- Choice matching is exact and case-sensitive. Integer ranges include both
  boundaries.
- Short options are not grouped; use `-a -b`, not `-ab`.
- A value following a text, integer, or list option is consumed even when it
  starts with `-`, so negative integers work as expected.
- Arguments after `--` are always positional.
- When subcommands are registered, the subcommand name is the first argument.

## Development

Run the source check and behavioral test matrix with:

```sh
KEIKA=keika sh scripts/test.sh
```

The test script emits Go into a temporary directory and removes it on exit.
It also builds an independent application that imports the library through its
public module path and a local replacement. Generated Go is not maintained as
source.

## License

Apache-2.0
