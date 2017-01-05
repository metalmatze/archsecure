package main

import "github.com/fatih/color"

type Status string

const (
	StatusVulnerable  Status = "Vulnerable"
	StatusTesting     Status = "Testing"
	StatusFixed       Status = "Fixed"
	StatusNotAffected Status = "Not affected"
)

func (s Status) Term() string {
	switch s {
	case StatusVulnerable:
		return color.RedString("%s", StatusVulnerable)
	case StatusTesting:
		return color.YellowString("%s", StatusTesting)
	case StatusFixed:
		return color.GreenString("%s", StatusFixed)
	default:
		return string(s)
	}
}
