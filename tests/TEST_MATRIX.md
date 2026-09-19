# Test matrix

The behavioral suite generates the Go package from `cli.km` and tests the
public contract rather than generated implementation details.

| Area | Normal cases | Errors and edges |
| --- | --- | --- |
| Boolean options | long, short, implicit true, every value accepted by `strconv.ParseBool` | empty and unsupported values |
| Text options | separate, inline, short, Unicode, empty, values beginning with `-` | missing value |
| Integer options | zero, sign, negative, platform minimum and maximum | empty, float, hexadecimal, whitespace and overflow |
| Choice options | default, required, long/short/inline, Unicode and explicitly allowed empty values | empty/duplicate choices, invalid defaults and unknown values |
| Integer ranges | default, required, negative values, exact/lower/upper boundaries | reversed ranges, defaults below/above range, input below/above range and non-integers |
| Repeatable options | absent, one/many, long/short/inline forms, ordering, empty and Unicode text, signed integers | required lists and invalid integers |
| Argument ordering | empty, positional, options before/after positionals | `--`, repeated `--`, grouped short options |
| Positional arity | exact zero, bounded/unbounded ranges and both boundaries | too few/many, invalid ranges, help/version bypass |
| Required options | present values, explicit empty text | missing values, help/version bypass |
| Help and version | long/short help, version, combined and repeated requests | unavailable version and unknown options |
| Definitions | valid command, long and short names | empty, whitespace, separators, reserved and duplicate names |
| Subcommands | direct, nested, full command paths, child options and child help | unknown, duplicate and invalid child commands |
| Result access | declared defaults and explicit values | unknown names return typed zero values |
| Executable examples | successful builds, basic output and advanced subcommand output | type/range failure, missing required option, diagnostics and exit status |
| External package | independent app installs through a v0.4.3 local replacement, imports `kinmokusei-cli`, then checks and builds | alias, package metadata, version and exported-source resolution failures stop the suite |
| Robustness | seeded and generated argument streams | arbitrary command, option and short-name strings must not panic |
