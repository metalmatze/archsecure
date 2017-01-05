package main

import "github.com/fatih/color"

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityHigh     Severity = "High"
	SeverityMedium   Severity = "Medium"
	SeverityLow      Severity = "Low"
)

func (s Severity) Term() string {
	switch s {
	case SeverityCritical:
		return color.RedString("%s", SeverityCritical)
	case SeverityHigh:
		return color.RedString("%s", SeverityHigh)
	case SeverityMedium:
		return color.YellowString("%s", SeverityMedium)
	case SeverityLow:
		return color.GreenString("%s", SeverityLow)
	default:
		return string(s)
	}
}
