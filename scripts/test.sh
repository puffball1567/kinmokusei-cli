#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT HUP INT TERM

keika_command=${KEIKA:-keika}

cd "$project_dir"
"$keika_command" deps lock
"$keika_command" check cli.km
"$keika_command" check examples/basic/main.km
"$keika_command" check examples/subcommands/main.km
"$keika_command" emit-go -package cli -o "$test_dir/cli.go" cli.km
cp tests/*_test.go "$test_dir/"

cd "$test_dir"
go mod init example.com/kinmokusei-cli-test >/dev/null
go test -coverprofile="$test_dir/coverage.out" ./...
go tool cover -func="$test_dir/coverage.out"
coverage_total=$(go tool cover -func="$test_dir/coverage.out" | awk '/^total:/ { print $3 }')
if [ "$coverage_total" != "100.0%" ]; then
  printf 'statement coverage is %s, want 100.0%%\n' "$coverage_total" >&2
  exit 1
fi
go test -run='^$' -fuzz='^FuzzParseNeverPanics$' -fuzztime=1s
go test -run='^$' -fuzz='^FuzzDefinitionValidationNeverPanics$' -fuzztime=1s

cd "$project_dir"
"$keika_command" build -o "$test_dir/greeter" examples/basic/main.km
"$keika_command" build -o "$test_dir/packages" examples/subcommands/main.km
actual_output=$($test_dir/greeter --name Kinmokusei --count 2)
expected_output=$(printf 'Hello, Kinmokusei!\nHello, Kinmokusei!')
if [ "$actual_output" != "$expected_output" ]; then
  printf 'unexpected example output:\n%s\n' "$actual_output" >&2
  exit 1
fi

actual_output=$($test_dir/packages add alpha beta --tag stable --tag=featured)
expected_output=$(printf 'add alpha\nadd beta\ntag stable\ntag featured')
if [ "$actual_output" != "$expected_output" ]; then
  printf 'unexpected subcommand example output:\n%s\n' "$actual_output" >&2
  exit 1
fi

help_output=$($test_dir/greeter --help)
case "$help_output" in
  *"Usage: greeter [options] [arguments]"*"--name <text>"*) ;;
  *)
    printf 'unexpected help output:\n%s\n' "$help_output" >&2
    exit 1
    ;;
esac

failure_output="$test_dir/failure.txt"
if "$test_dir/greeter" --count many >"$failure_output" 2>&1; then
  printf 'invalid integer unexpectedly succeeded\n' >&2
  exit 1
else
  failure_status=$?
fi
if [ "$failure_status" -ne 2 ] || ! grep -F 'expected integer' "$failure_output" >/dev/null; then
  printf 'unexpected invalid-integer failure (status %s):\n' "$failure_status" >&2
  cat "$failure_output" >&2
  exit 1
fi

if "$test_dir/greeter" --name Kinmokusei --count 0 >"$failure_output" 2>&1; then
  printf 'out-of-range integer unexpectedly succeeded\n' >&2
  exit 1
else
  failure_status=$?
fi
if [ "$failure_status" -ne 2 ] || ! grep -F 'expected 1..100' "$failure_output" >/dev/null; then
  printf 'unexpected integer-range failure (status %s):\n' "$failure_status" >&2
  cat "$failure_output" >&2
  exit 1
fi

if "$test_dir/greeter" --count 1 >"$failure_output" 2>&1; then
  printf 'missing required option unexpectedly succeeded\n' >&2
  exit 1
else
  failure_status=$?
fi
if [ "$failure_status" -ne 2 ] || ! grep -F 'required option --name is missing' "$failure_output" >/dev/null; then
  printf 'unexpected required-option failure (status %s):\n' "$failure_status" >&2
  cat "$failure_output" >&2
  exit 1
fi

library_dir="$test_dir/library"
consumer_dir="$test_dir/external-consumer"
mkdir -p "$library_dir"
cp cli.km go.mod kinmokusei.lock kinmokusei.toml "$library_dir/"
"$keika_command" new app --module example.com/kinmokusei-cli-consumer "$consumer_dir"
cp tests/external_consumer/main.km "$consumer_dir/main.km"
cd "$consumer_dir"
"$keika_command" deps add --offline --replace ../library github.com/puffball1567/kinmokusei-cli@v0.1.0
"$keika_command" check
"$keika_command" build -o "$test_dir/external-consumer-app"
actual_output=$($test_dir/external-consumer-app)
expected_output='format=json jobs=4'
if [ "$actual_output" != "$expected_output" ]; then
  printf 'unexpected external-consumer output:\n%s\n' "$actual_output" >&2
  exit 1
fi
