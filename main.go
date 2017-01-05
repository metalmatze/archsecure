package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"

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

	// Fetch all currently listed issues from Arch Linux's
	// security issues tracker.
	issues, err := ListIssues()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve all locally installed packages including
	// the specific version installed.
	installedPkgs, err := ListPackages()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nKNOWN ISSUES:\n-------------\n\n")
	for _, i := range issues {
		fmt.Printf("%+v\n", i)
	}
	fmt.Println()

	fmt.Printf("\nINSTALLED PKGS (OFFICIAL REPOSITORIES):\n---------------------------------------\n\n")
	for index, p := range installedPkgs {

		if index < 100 {
			fmt.Printf("%+v\n", p)
		} else if index == 100 {
			fmt.Printf("... (showing: 100, in total: %d) ...\n", len(installedPkgs))
		}
	}
	fmt.Println()
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
