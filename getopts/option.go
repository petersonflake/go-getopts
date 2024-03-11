//Package getopts parses command line options.  Allows --long-arguments, -s short arguments,
//flags to be set with -f or --long-option or --long-option=true; and
//un-set with +f or --long-option=false.
//
//Clumps of flags can be set with -fvx, un-set with +fvx.
//Short options that take arguments can be passed like
//-f file.txt or -ffile.txt.
//
//Long options taking an argument can be passed like
//--file=file.txt or --file file.txt.
//
//The argument '--' stops parsing and treats the rest of the argument vector
//as command arguments.  Each argument has a boolean indicating whether or not
//it comes after '--'.  This is to allow special handling of arguments like '-',
//which is usually used to read standard-input, but can also be the name of a file.
//Analogously, can be used for other arguments that may be commands, or file names.
package getopts

import "os"
import "fmt"
import "strings"

//This struct contains the argument passed
//and whether it was before or after '--'
type Rest struct {
	//What was passed
	Argument	string
	//Whether this argument comes after '--'
	AfterDashes	bool
}

//Command line options that take arguments.  Each subsequent occurence of the option
//overwrites OptArg and appends to OptArgs.
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

//Command line flag.  Stores the most recent value as boolean, and saves net count, so
//-f increases count and +f decreases it.
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

//Common information for options.
type option struct {
	ShortOpt	byte
	LongOpt		string
	Help		string
	Passed		bool
	takesArg	bool
}

//Assign value to flag, update count, and invoke event if applicable.
func (f *Flag)takeValue(value bool) {
	if value {
		f.Count++
	} else {
		f.Count--
	}
	f.Passed = value
	if value && f.OnTrue != nil {
		f.OnTrue()
	} else if !value && f.OnFalse != nil {
		f.OnFalse()
	}
}

//Whether this is an option that takes an argument -> true
//or a flag -> false
func (o option)takesArgument() bool {
	return o.takesArg
}

//Option or flag.  Exists mostly so they can be stored in same
//array with using 'any'.
type parameter interface {
	takesArgument() bool
}

//Add argument to Rest array.  If OnRestArg is not nil, we invoke it
//on the argument, and we add it to array only if that function returns
//true.  This is to support cases where the program interprets some sort
//of command language or similar.
func addRest(rest []Rest, arg string, dash bool) []Rest {
	if OnRestArg != nil {
		if OnRestArg(arg, dash) {
			rest = append(rest, Rest{
				Argument:	arg,
				AfterDashes:		dash,
			})
		}
	} else {
		rest = append(rest, Rest{
			Argument:	arg,
			AfterDashes:		dash,
		})
	}
	return rest
}

//Add option argument to optarg vector and invoke
//event if applicable.
func (o *Option)addOptArg(arg string) {
	o.OptArgs = append(o.OptArgs, arg)
	o.OptArg = arg
	o.Passed = true
	if o.Action != nil {
		o.Action(arg)
	}
}


var Options []Option = make([]Option, 0)

var Flags []Flag = make([]Flag, 0)

var paramsByShort map[byte]parameter = make(map[byte]parameter)

var paramsByLong map[string]parameter = make(map[string]parameter)

var OnRestArg func(arg string, afterDash bool) bool

func resetParams() {
	paramsByShort = make(map[byte]parameter)
	paramsByLong = make(map[string]parameter)
	Options = make([]Option, 0)
	Flags = make([]Flag, 0)
	OnRestArg = nil
}

func parseFlagOpt(flag, value string) (bool, error) {
	if strings.EqualFold(value, "f") {
		return false, nil
	} else if strings.EqualFold(value, "t") {
		return true, nil
	}

	if strings.EqualFold(value, "y") {
		return true, nil
	} else if strings.EqualFold(value, "n") {
		return false, nil
	}

	if strings.EqualFold(value, "no") {
		return false, nil
	} else if strings.EqualFold(value, "yes") {
		return true, nil
	}

	if strings.EqualFold(value, "true") {
		return true, nil
	} else if strings.EqualFold(value, "false") {
		return false, nil
	}

	return false, fmt.Errorf(errPassedOptargToFlag, flag)
}

//Ensure duplicate flags/options cannot be created
func checkShort(s byte) {
	if _, ok := paramsByShort[s]; ok {
		panic("Adding another command line option with same short option")
	}
}

func checkLong(l string) {
	if _, ok := paramsByLong[l]; ok {
		panic("Adding another command line option with same long option")
	}
}

func NewFlag(s byte, l string, h string) *Flag {
	checkShort(s)
	checkLong(l)

	flag := Flag{
		option:	option{
			ShortOpt:	s,
			LongOpt:	l,
			Help:		h,
			takesArg:	false,
		},
	}

	Flags = append(Flags, flag)
	paramsByShort[s] = &flag
	paramsByLong[l] = &flag
	return &flag
}

func NewFlagShort(s byte, h string) *Flag {
	checkShort(s)
	flag := Flag{
		option:	option{
			ShortOpt:	s,
			Help:		h,
			takesArg:	false,
		},
	}

	Flags = append(Flags, flag)
	paramsByShort[s] = &flag
	return &flag
}

func NewFlagLong(l string, h string) *Flag {
	checkLong(l)
	flag := Flag{
		option:	option{
			LongOpt:	l,
			Help:		h,
			takesArg:	false,
		},
	}

	Flags = append(Flags, flag)
	paramsByLong[l] = &flag
	return &flag
}

func NewOption(s byte, l string, h string) *Option {
	checkShort(s)
	checkLong(l)
	opt := Option{
		option: option{
			LongOpt:	l,
			ShortOpt:	s,
			Help:		h,
			takesArg:	true,
		},
	}

	Options = append(Options, opt)
	paramsByShort[s] = &opt
	paramsByLong[l] = &opt
	return &opt
}

func NewOptionShort(s byte, h string) *Option {
	checkShort(s)
	opt := Option{
		option: option{
			ShortOpt:	s,
			Help:		h,
			takesArg:	true,
		},
	}

	Options = append(Options, opt)
	paramsByShort[s] = &opt
	return &opt
}

func NewOptionLong(l string, h string) *Option {
	checkLong(l)
	opt := Option{
		option: option{
			LongOpt:	l,
			Help:		h,
			takesArg:	true,
		},
	}

	Options = append(Options, opt)
	paramsByLong[l] = &opt
	return &opt
}

const(
	errUnrecognizedShort = "Unrecognized short option:  %c"
	errUnrecognizedLong = "Unrecognized long option:  %s"
	errTriedToNegateOptArg = "Passed negation for option expecting argument: %c"
	errPassedOptargToFlag = "Passed non-boolean option to flag:  %s"
)

func ArgParse(argv []string) ([]Rest, error) {
	i := 1
	argc := len(argv)
	rest := make([]Rest, 0)
	expect_optarg := false
	var waiting_opt *Option
	for ; i < argc; i++ {
		arg := argv[i]
		if expect_optarg {
			waiting_opt.addOptArg(arg)
			expect_optarg = false
			continue
		}

		l := len(arg)
		switch l {
		case 0:		//Ignore empty arguments
		case 1: 	//Either '-' or an argument
			//rest = append(rest, arg)
			rest = addRest(rest, arg, false)
		case 2: 	//Either -a, +b, --, or rest
			if arg == "--" {
				for i++; i < argc; i++ {
					//rest = append(rest, arg)
					rest = addRest(rest, argv[i], true)
				}
				return rest, nil
			} else if arg[0] == '-' {
				if p, ok := paramsByShort[arg[1]]; ok {
					if p.takesArgument() {
						waiting_opt = p.(*Option)
						expect_optarg = true
					} else {
						p.(*Flag).takeValue(true)
					}
				} else {
					return rest, fmt.Errorf(errUnrecognizedShort, arg[1])
				}
			} else if arg[0] == '+' {
				if p, ok := paramsByShort[arg[1]]; ok {
					if p.takesArgument() {
						return rest, fmt.Errorf(errTriedToNegateOptArg, arg[1])
					} else {
						p.(*Flag).takeValue(false)
					}
				} else {
					return rest, fmt.Errorf(errUnrecognizedShort, arg[1])
				}
			} else {
				//rest = append(rest, arg)
				rest = addRest(rest, arg, false)

			}
		default:	//Either --blah or --foo=bar or -abc or +abc or rest
			if arg[0] == '-' {
				if arg[1] == '-' {
					//Long option
					indexOfEquals := strings.IndexByte(arg, '=')
					if indexOfEquals < 0 {
						long := arg[2:]
						if p, ok := paramsByLong[long]; ok {
							if p.takesArgument() {
								waiting_opt = p.(*Option)
								expect_optarg = true
							} else {
								p.(*Flag).takeValue(true)
							}
						} else {
							return rest, fmt.Errorf(errUnrecognizedLong, long)
						}
					} else {
						long := arg[2:indexOfEquals]
						optarg := arg[indexOfEquals+1:]
						if p, ok := paramsByLong[long]; ok {
							if p.takesArgument() {
								p.(*Option).addOptArg(optarg)
							} else {
								v, err := parseFlagOpt(long, optarg)
								if err != nil {
									return rest, err
								} else {
									p.(*Flag).takeValue(v)
								}
							}
						} else {
							return rest, fmt.Errorf(errUnrecognizedLong, long)
						}
					}
				} else {
					//clump
					for j := 1; j < len(arg); j++ {
						if p, ok := paramsByShort[arg[j]]; ok {
							if p.takesArgument() {
								if j < len(arg) - 1 {
									//The rest of the clump is the argument to last
									//recognized short option
									p.(*Option).addOptArg(arg[j:])
									break
								} else {
									//Here j == len(arg) - 1, index of last byte
									waiting_opt = p.(*Option)
									expect_optarg = true
								}
							} else {
								p.(*Flag).takeValue(true)
							}
						} else {
							return rest, fmt.Errorf(errUnrecognizedShort, arg[j])
						}
					}
				}
			} else if arg[0] == '+' {
				//Negate clump
				for j := 1; j < len(arg); j++ {
					if p, ok := paramsByShort[arg[j]]; ok {
						if p.takesArgument() {
							return rest, fmt.Errorf(errTriedToNegateOptArg, arg[j])
						} else {
							p.(*Flag).takeValue(false)
						}
					} else {
						return rest, fmt.Errorf(errUnrecognizedShort, arg[j])
					}
				}
			} else {
				//rest = append(rest, arg)
				rest = addRest(rest, arg, false)
			}
		}
	}

	return rest, nil
}

func GetOpts() ([]Rest, error) {
	return ArgParse(os.Args)
}

func ShowHelp() {
	panic("TODO")
}
