package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	issues, err := ListIssues()
	if err != nil {
		log.Println(err)
	}

	for _, i := range issues {
		fmt.Printf("%+v\n", i)
	}
}

type Issue struct {
	Group    string
	CVE      []string
	Package  []string
	Version  string
	Fixed    string
	Severity string
	Status   string
	Ticket   string
	Advisory string
}

func ListIssues() ([]Issue, error) {
	doc, err := goquery.NewDocument("https://security.archlinux.org/")
	if err != nil {
		return nil, err
	}

	var issues []Issue

	table := doc.Find(".content table tbody")
	table.Children().Each(func(i int, s *goquery.Selection) {
		issue := Issue{}
		s.Children().Each(func(j int, ss *goquery.Selection) {
			switch j {
			case 0:
				issue.Group = strings.TrimSpace(ss.Text())
			case 1:
				ss.Find("a").Each(func(k int, cveLink *goquery.Selection) {
					issue.CVE = append(issue.CVE, cveLink.Text())
				})
			case 2:
				ss.Find("a").Each(func(k int, packageLink *goquery.Selection) {
					issue.Package = append(issue.Package, packageLink.Text())
				})
			case 3:
				issue.Version = strings.TrimSpace(ss.Text())
			case 4:
				issue.Fixed = strings.TrimSpace(ss.Text())
			case 5:
				issue.Severity = strings.TrimSpace(ss.Text())
			case 6:
				issue.Status = strings.TrimSpace(ss.Text())
			case 7:
				issue.Ticket = strings.TrimSpace(ss.Text())
			case 8:
				issue.Advisory = strings.TrimSpace(ss.Text())
			}
		})

		issues = append(issues, issue)
	})

	return issues, nil
}
