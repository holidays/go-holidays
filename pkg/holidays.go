package holidays

import (
	"time"
)

type Holiday struct {
	Date    time.Time
	Name    string
	Regions []string
}

type Options struct {
	Informal bool
	Observed bool
}

func Between(start, end time.Time, regions []string, options Options) (holidays []Holiday, err error) {
	return []Holiday{}, nil
}

func AvailableRegions() ([]string, error) {
	return []string{}, nil
}
