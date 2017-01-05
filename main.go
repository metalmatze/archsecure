package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	"github.com/PuerkitoBio/goquery"
)

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

type Package struct {
	Name    string
	Version string
}

func main() {
	issues, err := ListIssues()
	if err != nil {
		log.Fatal(err)
	}

	installedPkgs, err := ListPackages()
	if err != nil {
		log.Fatal(err)
	}

	const padding = 3
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.TabIndent)

	fmt.Fprintln(w, "Package\tVersion\tFixed\tSeverity\tStatus\tGroup")

	for _, pkg := range installedPkgs {
		for _, issue := range issues {
			for _, issuePkg := range issue.Package {
				if pkg.Name == issuePkg && pkg.Version == issue.Version {
					fmt.Fprintf(w,
						"%s\t%s\t%s\t%s\t%s\t%s\n",
						pkg.Name,
						pkg.Version,
						issue.Fixed,
						issue.Severity,
						issue.Status,
						issue.Group,
					)
				}
			}
		}
	}
	w.Flush()
}

func ListIssues() ([]Issue, error) {
	// Fetch content of security tracker website and
	// make it available to goquery parser.
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

func ListPackages() ([]Package, error) {
	var buf bytes.Buffer

	// Prepare a pacman query to retrieve installed packages.
	cmd := exec.Command("pacman", "-Qs")
	cmd.Stdout = &buf

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Trim possibly leading and trailing space.
	pkgsListRaw := strings.TrimSpace(buf.String())

	// Split at newline.
	pkgsList := strings.Split(pkgsListRaw, "\n")

	pkgs := make([]Package, 0, len(pkgsList)/2)

	for _, pkgLine := range pkgsList {
		if !strings.HasPrefix(pkgLine, "    ") {

			// This is not a package's description line because
			// it does not begin with an indentation. Proceed.

			// Split at space to remove additional information.
			pkgData := strings.Split(pkgLine, " ")

			// Again, split first part of package name to remove
			// its membership to package repositories.
			pkgName := strings.Split(pkgData[0], "/")

			pkg := Package{
				Name:    pkgName[1],
				Version: pkgData[1],
			}

			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}
