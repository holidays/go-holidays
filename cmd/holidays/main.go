// holidays is a small command-line wrapper around the go-holidays library.
//
// Subcommands:
//
//	holidays on DATE         [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
//	holidays between A B     [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
//	holidays year YYYY       [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
//	holidays next N FROM     [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
//	holidays workweek DATE   [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
//	holidays regions
//
// Dates are YYYY-M-D (single- or double-digit month/day both accepted).
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	holidays "github.com/holidays/go-holidays"
)

// osExit is os.Exit by default; tests swap it out so a failing run() can be
// observed without terminating the test binary.
var osExit = os.Exit

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "holidays:", err)
		osExit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return fmt.Errorf("missing subcommand")
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "on":
		return cmdOn(rest)
	case "between":
		return cmdBetween(rest)
	case "year":
		return cmdYear(rest)
	case "next":
		return cmdNext(rest)
	case "workweek":
		return cmdWorkweek(rest)
	case "regions":
		return cmdRegions(rest)
	case "-h", "--help", "help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown subcommand %q", cmd)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `usage:
  holidays on DATE         [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
  holidays between A B     [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
  holidays year YYYY       [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
  holidays next N FROM     [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
  holidays workweek DATE   [--regions r1,r2] [--informal] [--observed] [--custom f1,f2]
  holidays regions`)
}

func cmdOn(args []string) error {
	fs := newFlags("on")
	cf := bindCommonFlags(fs)
	if err := parseLeading(fs, args, 1); err != nil {
		return err
	}
	d, err := parseDate(fs.Arg(0))
	if err != nil {
		return err
	}
	opts, err := cf.options()
	if err != nil {
		return err
	}
	hs, err := holidays.On(d, opts)
	if err != nil {
		return err
	}
	return printHolidays(hs)
}

func cmdBetween(args []string) error {
	fs := newFlags("between")
	cf := bindCommonFlags(fs)
	if err := parseLeading(fs, args, 2); err != nil {
		return err
	}
	start, err := parseDate(fs.Arg(0))
	if err != nil {
		return err
	}
	end, err := parseDate(fs.Arg(1))
	if err != nil {
		return err
	}
	opts, err := cf.options()
	if err != nil {
		return err
	}
	hs, err := holidays.Between(start, end, opts)
	if err != nil {
		return err
	}
	return printHolidays(hs)
}

func cmdYear(args []string) error {
	fs := newFlags("year")
	cf := bindCommonFlags(fs)
	if err := parseLeading(fs, args, 1); err != nil {
		return err
	}
	year, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("year: %w", err)
	}
	opts, err := cf.options()
	if err != nil {
		return err
	}
	hs, err := holidays.YearHolidays(year, opts)
	if err != nil {
		return err
	}
	return printHolidays(hs)
}

func cmdNext(args []string) error {
	fs := newFlags("next")
	cf := bindCommonFlags(fs)
	if err := parseLeading(fs, args, 2); err != nil {
		return err
	}
	count, err := strconv.Atoi(fs.Arg(0))
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}
	from, err := parseDate(fs.Arg(1))
	if err != nil {
		return err
	}
	opts, err := cf.options()
	if err != nil {
		return err
	}
	hs, err := holidays.NextHolidays(from, count, opts)
	if err != nil {
		return err
	}
	return printHolidays(hs)
}

func cmdWorkweek(args []string) error {
	fs := newFlags("workweek")
	cf := bindCommonFlags(fs)
	if err := parseLeading(fs, args, 1); err != nil {
		return err
	}
	d, err := parseDate(fs.Arg(0))
	if err != nil {
		return err
	}
	opts, err := cf.options()
	if err != nil {
		return err
	}
	any, err := holidays.AnyHolidaysDuringWorkWeek(d, opts)
	if err != nil {
		return err
	}
	fmt.Println(any)
	return nil
}

func cmdRegions(args []string) error {
	fs := newFlags("regions")
	if err := parseLeading(fs, args, 0); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("regions takes no arguments")
	}
	for _, r := range holidays.AvailableRegions() {
		fmt.Println(r)
	}
	return nil
}

func newFlags(name string) *flag.FlagSet {
	return flag.NewFlagSet(name, flag.ContinueOnError)
}

// commonFlags holds the flags shared by every subcommand that queries holidays.
type commonFlags struct {
	regions  *string
	informal *bool
	observed *bool
	custom   *string
}

func bindCommonFlags(fs *flag.FlagSet) *commonFlags {
	return &commonFlags{
		regions:  fs.String("regions", "", "comma-separated region codes (empty == all)"),
		informal: fs.Bool("informal", false, "include informal holidays"),
		observed: fs.Bool("observed", false, "use observed dates"),
		custom:   fs.String("custom", "", "comma-separated custom definition files to load first"),
	}
}

// options loads any --custom definition files and returns the library options
// described by the parsed flags.
func (cf *commonFlags) options() (holidays.Options, error) {
	if paths := splitList(*cf.custom); len(paths) > 0 {
		if err := holidays.LoadCustom(paths...); err != nil {
			return holidays.Options{}, fmt.Errorf("custom: %w", err)
		}
	}
	return holidays.Options{
		Regions:  splitList(*cf.regions),
		Informal: *cf.informal,
		Observed: *cf.observed,
	}, nil
}

func parseLeading(fs *flag.FlagSet, args []string, n int) error {
	if err := fs.Parse(reorderFlagsFirst(args)); err != nil {
		return err
	}
	if fs.NArg() < n {
		return fmt.Errorf("%s expects %d positional argument(s)", fs.Name(), n)
	}
	return nil
}

// reorderFlagsFirst lets users put flags after positional args (e.g.
// `year 2024 --regions us`) by moving flags to the front before flag.Parse.
// The reordered positionals are placed after a "--" terminator so flag.Parse
// treats them as operands: without it a positional that looks like a flag
// (a bare negative integer such as "-3") would be rejected as undefined.
var boolFlagNames = map[string]bool{"informal": true, "observed": true}

func reorderFlagsFirst(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") || isNegativeInt(a) {
			positional = append(positional, a)
			continue
		}
		flags = append(flags, a)
		if strings.Contains(a, "=") {
			continue
		}
		name := strings.TrimLeft(a, "-")
		if boolFlagNames[name] {
			continue
		}
		if i+1 < len(args) {
			flags = append(flags, args[i+1])
			i++
		}
	}
	if len(positional) == 0 {
		return flags
	}
	return append(flags, append([]string{"--"}, positional...)...)
}

// isNegativeInt reports whether s is a bare negative integer (for example
// "-3"). Such a token is a positional argument (a count or year), not a flag,
// so reorderFlagsFirst must not route it into the flags bucket where
// flag.Parse would reject it as an undefined flag.
func isNegativeInt(s string) bool {
	if len(s) < 2 || s[0] != '-' {
		return false
	}
	for _, r := range s[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-1-2", s)
}

func printHolidays(hs []holidays.Holiday) error {
	sort.Slice(hs, func(i, j int) bool {
		if !hs[i].Date.Equal(hs[j].Date) {
			return hs[i].Date.Before(hs[j].Date)
		}
		return hs[i].Name < hs[j].Name
	})
	for _, h := range hs {
		fmt.Printf("%s\t%s\t%s\n", h.Date.Format("2006-01-02"), h.Name, strings.Join(h.Regions, ","))
	}
	return nil
}
