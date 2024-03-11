# go-getopts
## Introduction
`go-getopts` is a simple library for parsing command-line arguments.  Supports common
conventions like being able to negate an option with `+`.  Short options can
be passed as clumps like `-vxf` or negated with `+vxf`.  Long options for
flags can be passed as `--verbose` or `--verbose=true` and negated like
`--verbose=false`.  Options that take an argument can be passed like
`--file file.txt` or `--file=file.txt` or `-f file.txt` or `-ffile.txt`.

Command line arguments that do not start with `-` are grouped together and
returned by the GetOpts function.

When parsing a clump of options, if preceded by a dash, each flag is set to
true.  When an option that takes an argument is encountered, the entire rest
of the clump is set to be the option argument.

If the clump is being negated then all flags are set to false.  An option that
takes an argument in the clump will result in an error.

If one of the arguments is `--`, then all arguments after that one are
combined into a single array.

## Types
### Flag
Used to represent boolean options and counts.

Example:  If we have an option with the short form `v` and the long form
`verbose`, then we can call `<program> -v` or `<program> --verbose` or
`<program> --verbose=true` and all of these will set `verbose` to true and
increment the count of that option by one.

We can also negate options.  So `<program> +v` and `<program> --verbose=false`
will both set the option to false and decrment its count.

```go
type Flag struct {
	//Contains information common to options and flags.
	option
	//The number of times the flag was passed minus the number of times
	//the flag was negated.  Used for verbosity, debug-level, etc.
	Count	int
	//If present, this function is called each time the flag is passed
	OnTrue	func()
	//If present, function called each time flag is negated
	//by +f or --flag=false
	OnFalse	func()
}

```

### Option
Used for options that take an argument.  This type stores the most recent
option argument in Optarg, and stores all previous option arguments in an
array.

When we want all arguments passed to an option, read
`Option.Optargs`.  The latter is useful for cases where, for example, we want
to pass multiple input files or config files to a program.

```go
type Option struct {
	//Contains information common to options and flags.
	option
	//The most recent opt-arg for this option.  Used for options where only
	//one value would be relevant so the user can override previous values.  This
	//is useful for scripts and aliases.
	OptArg	string
	//All opt-args that were passed to this option.  Used for cases like passing files
	//to process.
	OptArgs	[]string
	//If present, this function is called with the opt-arg as an argument as soon as it
	//is parsed.
	Action	func(string)
}

```

### Rest
An argument passed to the program that was not a flag or option.  For input
files, etc.  The boolean member `AfterDashes` was added to handle the common
convention where `-` means process standard input before `--`, and a file
called `-` when passed after.

```go
type Rest struct {
	//What was passed
	Argument	string
	//Whether this argument comes after '--'
	AfterDashes	bool
}

```
