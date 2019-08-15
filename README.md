# logschema

A tool for making sure log entries fulfill a given json schema.

## Installation

```bash
go get github.com/Typeform/logschema/...
go install github.com/Typeform/logschema/cmd/logschema
```

## Usage

By default, `logschema` finds a logschema.json file in the executing directory.
To pass a location for the schema, use `--schema=/path/to/logschema.json`

### By passing a log file

```bash
logschema myfile.log
```

### By piping a stream

```bash
tail -f myfile.log | logschema
```
