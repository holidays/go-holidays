# go-holidays

A Go library for working with statutory and other holidays.

All holiday definitions are maintained in the
[holidays/definitions](https://github.com/holidays/definitions) repository
(vendored here as a submodule). By default this library returns statutory
(formally government-defined) holidays. Culturally recognized but non-statutory
holidays (such as Valentine's Day) are available via the `Informal` option. See
the [definitions syntax guide](https://github.com/holidays/definitions/blob/master/doc/SYNTAX.md#formalinformal)
for details on how holidays are classified.

## Installation

```bash
go get github.com/holidays/go-holidays
```

## Tested versions

This module requires **Go 1.25+**, the floor declared by the `go` directive in
`go.mod`.

CI runs the full test suite against the latest Go minor release and the
previous two:

  * 1.27
  * 1.26
  * 1.25

That list lives in the `test-matrix` job in `.github/workflows/ci.yml` and moves
forward as new Go minors are released, with the floor in `go.mod` moving up in
step so every matrix entry is actually exercised (rather than silently
upgraded by `GOTOOLCHAIN=auto`).

## Semver

This module follows [semantic versioning](http://semver.org/). The guarantee
specifically covers the exported surface of the root package
`github.com/holidays/go-holidays` (its functions, types, and their fields).

Please note that we consider definition changes to be "minor" bumps, meaning
they are backwards compatible with your code but might give different holiday
results.

## Time zones

Dates are always constructed in UTC and truncated to the calendar day
(`time.Date(y, m, d, 0, 0, 0, 0, time.UTC)`). Pass in whatever time zone you
like: comparisons are done on the UTC calendar day, not on wall-clock time.

## Usage

Every example below assumes this import, shown once here and omitted from the
rest for brevity:

```go
import (
	"time"

	holidays "github.com/holidays/go-holidays"
)
```

This library offers multiple ways to check for holidays for a variety of
scenarios.

#### Checking a specific date

Get all holidays on April 25, 2008 in Australia:

```go
d := time.Date(2008, time.April, 25, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{Regions: []string{"au"}})
// hs[0].Name == "ANZAC Day"
```

You can check multiple regions in a single call:

```go
d := time.Date(2008, time.January, 1, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{Regions: []string{"us", "fr"}})
// hs contains "New Year's Day" (regions: [us]) and "Jour de l'an" (regions: [fr])
```

You can leave `Regions` empty to get holidays for any registered region:

```go
d := time.Date(2007, time.April, 25, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{})
// hs contains "ANZAC Day" (au), "Festa della Liberazione" (it), ...
```

#### Wildcard regions

A region code ending in an underscore (e.g. `au_`, `ca_`) is a *wildcard*. It
matches the parent country region and all of its sub-regions in a single call:

```go
d := time.Date(2017, time.March, 13, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{Regions: []string{"au_"}})
// hs contains "Eight Hours Day" (au_tas), "Labour Day" (au_vic),
// "March Public Holiday" (au_sa), "Canberra Day" (au_act)
```

The same date queried with the plain `au` region returns nothing, because none
of those holidays are observed nation-wide:

```go
hs, err := holidays.On(d, holidays.Options{Regions: []string{"au"}})
// hs is empty
```

Use a wildcard when you want "this country and every sub-region it defines"
without listing each sub-region explicitly.

Note that a wildcard always collapses to the top-level country region. The
portion between the country prefix and the trailing underscore is ignored, so
`au_vic_` behaves identically to `au_` (it loads every Australian sub-region,
not just Victoria's). There is currently no way to wildcard-match only the
children of a sub-region.

#### Checking a date range

Get all holidays during the month of July 2008 in Canada and the US:

```go
from := time.Date(2008, time.July, 1, 0, 0, 0, 0, time.UTC)
to := time.Date(2008, time.July, 31, 0, 0, 0, 0, time.UTC)
hs, err := holidays.Between(from, to, holidays.Options{Regions: []string{"ca", "us"}})
// hs contains "Canada Day", "Independence Day"
```

#### Informal holidays

Set `Options.Informal` to include holidays specified as informal in your
results. See the [definitions syntax guide](https://github.com/holidays/definitions/blob/master/doc/SYNTAX.md#formalinformal)
for what constitutes "informal" vs "formal".

By default this option is `false`, meaning no informal holidays are returned.

Get Valentine's Day in the US:

```go
d := time.Date(2018, time.February, 14, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{Regions: []string{"us"}, Informal: true})
// hs[0].Name == "Valentine's Day"
```

Leaving `Informal` false means Valentine's Day is not returned:

```go
hs, err := holidays.On(d, holidays.Options{Regions: []string{"us"}})
// hs is empty
```

#### Observed holidays

Set `Options.Observed` to include holidays that are observed on different days
than they actually occur. See the [definitions syntax guide](https://github.com/holidays/definitions/blob/master/doc/SYNTAX.md#observed)
for further explanation of "observed".

By default this option is `false`, meaning no observed logic is applied.

Get holidays that are observed on Monday July 2, 2007 in British Columbia,
Canada:

```go
d := time.Date(2007, time.July, 2, 0, 0, 0, 0, time.UTC)
hs, err := holidays.On(d, holidays.Options{Regions: []string{"ca_bc"}, Observed: true})
// hs[0].Name == "Canada Day"
```

Leaving `Observed` false means "Canada Day" is not returned on July 2, since it
actually falls on Sunday July 1:

```go
hs, err := holidays.On(d, holidays.Options{Regions: []string{"ca_bc"}})
// hs is empty

d = time.Date(2007, time.July, 1, 0, 0, 0, 0, time.UTC)
hs, err = holidays.On(d, holidays.Options{Regions: []string{"ca_bc"}})
// hs[0].Name == "Canada Day"
```

#### Any holidays during work week

`AnyHolidaysDuringWorkWeek` reports whether any holiday matching the options
falls during the Monday-Friday work week containing the given date.

Check whether a holiday falls during the first week of the year for any
region:

```go
d := time.Date(2016, time.January, 1, 0, 0, 0, 0, time.UTC)
any, err := holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{})
// any == true
```

`Informal` and `Observed` apply the same way they do everywhere else:

```go
// true: Valentine's Day falls on a Wednesday
d = time.Date(2018, time.February, 14, 0, 0, 0, 0, time.UTC)
any, _ = holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{Regions: []string{"us"}, Informal: true})

// false without Informal
any, _ = holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{Regions: []string{"us"}})

// true: Veterans Day is observed on Monday November 12, 2018
d = time.Date(2018, time.November, 12, 0, 0, 0, 0, time.UTC)
any, _ = holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{Regions: []string{"us"}, Observed: true})

// false without Observed: the actual holiday is on Sunday November 11
any, _ = holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{Regions: []string{"us"}})
```

#### Next holidays

`NextHolidays` returns the next `count` holidays on or after a given date,
sorted by date ascending:

```go
from := time.Date(2016, time.February, 23, 0, 0, 0, 0, time.UTC)
hs, err := holidays.NextHolidays(from, 3, holidays.Options{Regions: []string{"us"}, Informal: true})
// hs contains "St. Patrick's Day", "Good Friday", "Easter Sunday"
```

#### Year holidays

`YearHolidays` returns every holiday matching the options in a given calendar
year:

```go
hs, err := holidays.YearHolidays(2016, holidays.Options{Regions: []string{"ca_on"}})
```

`YearHolidaysFrom` returns every holiday from a given date through December 31
of that date's year, sorted ascending:

```go
from := time.Date(2016, time.February, 23, 0, 0, 0, 0, time.UTC)
hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"ca_on"}})
// hs contains "Good Friday", "Easter Sunday", "Victoria Day", "Canada Day",
// "Civic Holiday", "Labour Day", "Thanksgiving", "Remembrance Day",
// "Christmas Day", "Boxing Day"
```

#### Available regions

`AvailableRegions` returns every registered region code, sorted
lexicographically:

```go
regions := holidays.AvailableRegions()
// regions == []string{"ar", "at", ..., "sg", ...}
```

#### Region display names

`RegionName` returns the display name for one region code, and whether it is
registered. `RegionNames` returns every registered region code mapped to its
display name:

```go
name, ok := holidays.RegionName("gb_sct")
// name == "Scotland", ok == true

names := holidays.RegionNames()
// names["ch_ge"] == "Genève"
```

## Command-line interface

`make build` produces `bin/holidays` and `bin/gen-holidays`. `bin/holidays`
wraps the same public API from the command line:

```bash
bin/holidays on 2024-07-04 --regions us
bin/holidays between 2024-12-20 2024-12-31 --regions us
bin/holidays year 2024 --regions us
bin/holidays next 5 2024-05-28 --regions us
bin/holidays workweek 2024-11-25 --regions us
bin/holidays regions
```

Every subcommand except `regions` also accepts `--informal` and `--observed`.
Flags may appear before or after the positional arguments.

## Loading custom definitions on the fly

In addition to the [provided definitions](https://github.com/holidays/definitions)
you can load a custom definitions file on the fly and use it immediately.

To load a custom "Company Founding" holiday on June 1st, put this in
`custom_holidays.yaml`:

```yaml
months:
  6:
  - name: Company Founding
    regions: [my_custom_region]
    mday: 1
```

Then load it and query it by the region code from its `regions:` list:

```go
err := holidays.LoadCustom("/home/user/holiday_definitions/custom_holidays.yaml")
hs, err := holidays.On(time.Date(2013, time.June, 1, 0, 0, 0, 0, time.UTC),
    holidays.Options{Regions: []string{"my_custom_region"}})
// hs[0].Name == "Company Founding"
```

Custom definition files must match the [syntax of the existing definition files](https://github.com/holidays/definitions/blob/master/doc/SYNTAX.md).
Region codes are lowercased and trimmed, the same as `Options.Regions`.

Multiple files can be loaded at the same time:

```go
err := holidays.LoadCustom(
    "/home/user/holidays/custom_holidays1.yaml",
    "/home/user/holidays/custom_holidays2.yaml",
)
```

Each file is keyed by its base name without the directory or extension. Loading
the same path again replaces its prior load, and so does loading any other file
with the same base name (`/a/holidays.yaml` and `/b/holidays.yaml`, or
`holidays.yaml` and `holidays.yml`), so give every file a distinct base name.
Files with different base names add rules without overwriting one another.

`LoadCustom` is all-or-nothing: if any file fails to read, parse, or validate,
it returns an error and none of the files are registered. `UnloadCustom`
removes rules previously loaded from the given paths, matched by base name.
Both call `ResetCache` when they succeed, since the rule set changed.

### Custom methods

Custom rules can't embed executable logic in YAML: a rule's `function:` or
`observed:` reference must point at a method already registered in Go before
you call `LoadCustom`, or `LoadCustom` returns an error:

```go
holidays.RegisterMethod("my_method", func(a holidays.MethodArgs) (time.Time, error) {
    return time.Date(a.Year, time.March, 15, 0, 0, 0, 0, time.UTC), nil
})
err := holidays.LoadCustom("my_team.yaml") // YAML can now use function: my_method(year)
```

The method receives only a `MethodArgs`. The parentheses in the YAML call are
required, but any arguments inside them (the `year` above) are not passed to
your function.

- `function:` methods return the holiday's date for `a.Year`. Only the month and
  day of the result are kept, after the rule's `function_modifier:` days are
  added: the year is forced to the year being resolved and the time zone to UTC.
  Returning the zero `time.Time` means the holiday does not occur that year.
  `a.Month` and `a.Day` are the rule's month and mday, `a.Date` is the date
  computed from them (or the rule's wday/week), and `a.Region` is the first
  region listed on the rule.
- `observed:` methods run only when `Options.Observed` is true. `a.Date` is the
  holiday's date, `a.Region` is the first region in `Options.Regions` (empty if
  none), and the date you return replaces the holiday's date.
- `RegisterMethod` panics if the name is already registered, including any
  built-in method name, and a registered method can't be removed. Register each
  name once, for example from an `init` function.

## Caching holiday lookups

If you are checking holidays regularly you can cache your results for improved
performance. Run this before looking up a holiday (e.g. during startup):

```go
year := 365 * 24 * time.Hour
err := holidays.CacheBetween(time.Now(), time.Now().Add(2*year), holidays.Options{
    Regions:  []string{"ca", "us"},
    Observed: true,
})
```

Holidays for the regions and options specified within the dates specified will
be pre-computed and stored in memory. Subsequent `On`/`Between` calls with the
same options and a range fully contained in the cached window are answered
from the cache. `ResetCache` clears every cached range.

## How to contribute

See [CONTRIBUTING.md](CONTRIBUTING.md) for information on how to help out.

## Credits

* Holiday definitions come from [holidays/definitions](https://github.com/holidays/definitions),
  along with all of its wonderful contributors.
