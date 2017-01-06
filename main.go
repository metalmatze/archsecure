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

type (
	Link struct {
		URL  string
		Text string
	}

	Issue struct {
		Group    Link
		CVE      []Link
		Package  []Link
		Version  string
		Fixed    string
		Severity Severity
		Status   Status
		Ticket   Link
		Advisory Link
	}

	Package struct {
		Name    string
		Version string
	}
)

const (
	URL = "https://security.archlinux.org"
)

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

	fmt.Fprintln(w, "PACKAGE\tVERSION\tFIXED\tSEVERITY\tSTATUS\tGROUP\tLINK")

	for _, pkg := range installedPkgs {
		for _, issue := range issues {
			for _, issuePkg := range issue.Package {
				if pkg.Name == issuePkg.Text && pkg.Version == issue.Version {
					fmt.Fprintf(w,
						"%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
						pkg.Name,
						pkg.Version,
						issue.Fixed,
						issue.Severity.Term(),
						issue.Status.Term(),
						issue.Group.Text,
						URL+issue.Group.URL,
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
	doc, err := goquery.NewDocument(URL)
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
				issue.Group = SelectionToLink(ss.Find("a"))
			case 1:
				ss.Find("a").Each(func(k int, cveLink *goquery.Selection) {
					issue.CVE = append(issue.CVE, SelectionToLink(cveLink))
				})
			case 2:
				ss.Find("a").Each(func(k int, pkgLink *goquery.Selection) {
					issue.Package = append(issue.Package, SelectionToLink(pkgLink))
				})
			case 3:
				issue.Version = strings.TrimSpace(ss.Text())
			case 4:
				issue.Fixed = strings.TrimSpace(ss.Text())
			case 5:
				issue.Severity = Severity(strings.TrimSpace(ss.Text()))
			case 6:
				issue.Status = Status(strings.TrimSpace(ss.Text()))
			case 7:
				issue.Ticket = SelectionToLink(ss.Find("a"))
			case 8:
				issue.Advisory = SelectionToLink(ss.Find("a"))
			}
		})

		issues = append(issues, issue)
	})

	return issues, nil
}

func SelectionToLink(s *goquery.Selection) Link {
	href, _ := s.Attr("href")

	return Link{
		URL:  href,
		Text: strings.TrimSpace(s.Text()),
	}
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
