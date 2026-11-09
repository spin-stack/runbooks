// Command check holds every runbook under runbooks/ to the shape the app reads: its front matter,
// its sections in order, no raw HTML, links to runbooks that exist, and nothing a public
// repository must not carry. It prints every problem and exits 1 when there is one.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

func main() {
	problems, err := check(os.DirFS("."), "runbooks")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, p := range problems {
		fmt.Println(p)
	}
	if len(problems) > 0 {
		os.Exit(1)
	}
}

// Sections is every section a runbook has, in the order it has them: what is seen, who it hurts,
// how to find out, what to do, how to know it worked, and whom to call when it did not.
var Sections = []string{"Symptoms", "Impact", "Diagnose", "Fix", "Verify", "Escalate"}

// Severities is what a runbook's severity may be: the alert rules' two, and info for a procedure
// nothing pages for.
var Severities = []string{"critical", "warning", "info"}

// AppliesTo is what sends an operator to a runbook. The app links each from where these are shown:
// an alert rule by its uid, a job by its name, an owner's notice by its event (HELD), a Grafana
// dashboard by its uid.
type AppliesTo struct {
	Alerts     []string `yaml:"alerts"`
	Jobs       []string `yaml:"jobs"`
	Events     []string `yaml:"events"`
	Dashboards []string `yaml:"dashboards"`
}

// Front is a runbook's front matter.
type Front struct {
	ID           string    `yaml:"id"`
	Title        string    `yaml:"title"`
	AppliesTo    AppliesTo `yaml:"applies_to"`
	Severity     string    `yaml:"severity"`
	Owner        string    `yaml:"owner"`
	LastReviewed string    `yaml:"last_reviewed"`
}

var (
	kebab = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	// What a public repository must not carry. An AWS account id is twelve digits on their own; a
	// private key announces itself; an installation's hostname is its own, and a runbook says
	// <your domain>.
	forbidden = []struct {
		re   *regexp.Regexp
		what string
	}{
		{regexp.MustCompile(`(^|[^0-9])[0-9]{12}([^0-9]|$)`), "an AWS account id"},
		{regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`), "a private key"},
		{regexp.MustCompile(`(?i)\bspinbox\.dev\b`), "an installation's own hostname (write <your domain>)"},
		{regexp.MustCompile(`\b(AKIA|ASIA)[A-Z0-9]{16}\b`), "an AWS access key id"},
	}
)

// check reads every runbook under dir and returns what is wrong with them, file by file.
func check(fsys fs.FS, dir string) ([]string, error) {
	names, err := fs.Glob(fsys, dir+"/*.md")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no runbooks under %s/", dir)
	}
	sort.Strings(names)
	var problems []string
	ids := map[string]string{}
	type link struct{ from, to string }
	var links []link
	for _, name := range names {
		raw, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, err
		}
		file := strings.TrimSuffix(path.Base(name), ".md")
		front, out, targets := one(file, raw)
		for _, p := range out {
			problems = append(problems, name+": "+p)
		}
		if other, ok := ids[front.ID]; ok && front.ID != "" {
			problems = append(problems, fmt.Sprintf("%s: id %q is used by %s too", name, front.ID, other))
		} else if front.ID != "" {
			ids[front.ID] = name
		}
		for _, t := range targets {
			links = append(links, link{name, t})
		}
	}
	for _, l := range links {
		if _, ok := ids[l.to]; !ok {
			problems = append(problems, fmt.Sprintf("%s: links runbook %q, which does not exist", l.from, l.to))
		}
	}
	return problems, nil
}

// one checks a runbook on its own and returns its front matter, its problems and the ids of the
// runbooks it links.
func one(file string, raw []byte) (Front, []string, []string) {
	var problems []string
	front, src, err := split(raw)
	if err != nil {
		return Front{}, []string{err.Error()}, nil
	}
	var f Front
	dec := yaml.NewDecoder(bytes.NewReader(front))
	dec.KnownFields(true)
	if err := dec.Decode(&f); err != nil {
		return Front{}, []string{"front matter: " + err.Error()}, nil
	}
	problems = append(problems, f.problems(file)...)

	for _, rule := range forbidden {
		if rule.re.Match(raw) {
			problems = append(problems, "carries "+rule.what)
		}
	}

	doc := goldmark.New(goldmark.WithExtensions(extension.GFM)).Parser().Parse(text.NewReader(src))
	var headings []string
	var targets []string
	err = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Heading:
			title := string(n.Lines().Value(src))
			switch {
			case n.Level == 1:
				problems = append(problems, "has a level-one heading: the title is the front matter's")
			case n.Level == 2:
				headings = append(headings, strings.TrimSpace(title))
			}
		case *ast.HTMLBlock, *ast.RawHTML:
			problems = append(problems, "carries raw HTML, which the app does not render")
		case *ast.Link:
			if t, ok := runbookLink(string(n.Destination)); ok {
				targets = append(targets, t)
			} else if p := badLink(string(n.Destination)); p != "" {
				problems = append(problems, p)
			}
		case *ast.AutoLink:
			if p := badLink(string(n.URL(src))); p != "" {
				problems = append(problems, p)
			}
		case *ast.Image:
			problems = append(problems, "has an image: the app renders text")
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		problems = append(problems, err.Error())
	}
	if !slices.Equal(headings, Sections) {
		problems = append(problems, fmt.Sprintf("sections are %q, not %q", headings, Sections))
	}
	return f, problems, targets
}

func (f Front) problems(file string) []string {
	var p []string
	switch {
	case f.ID == "":
		p = append(p, "has no id")
	case !kebab.MatchString(f.ID):
		p = append(p, fmt.Sprintf("id %q is not kebab case", f.ID))
	case f.ID != file:
		p = append(p, fmt.Sprintf("id %q is not the file's name %q", f.ID, file))
	}
	if strings.TrimSpace(f.Title) == "" {
		p = append(p, "has no title")
	}
	if !slices.Contains(Severities, f.Severity) {
		p = append(p, fmt.Sprintf("severity %q is none of %q", f.Severity, Severities))
	}
	if !kebab.MatchString(f.Owner) {
		p = append(p, fmt.Sprintf("owner %q is not a kebab-case role", f.Owner))
	}
	if _, err := time.Parse(time.DateOnly, f.LastReviewed); err != nil {
		p = append(p, fmt.Sprintf("last_reviewed %q is not a date (YYYY-MM-DD)", f.LastReviewed))
	}
	a := f.AppliesTo
	if len(a.Alerts)+len(a.Jobs)+len(a.Events)+len(a.Dashboards) == 0 && f.Severity != "info" {
		p = append(p, "applies to nothing: a runbook for an alert names it, and a procedure is info")
	}
	for _, list := range [][]string{a.Alerts, a.Jobs, a.Dashboards} {
		for _, s := range list {
			if !kebab.MatchString(s) {
				p = append(p, fmt.Sprintf("applies_to %q is not kebab case", s))
			}
		}
	}
	for _, e := range a.Events {
		if e != strings.ToUpper(e) || strings.HasPrefix(e, "WORKSPACE_EVENT_TYPE_") {
			p = append(p, fmt.Sprintf("event %q is not written as the app names it (HELD)", e))
		}
	}
	return p
}

// split cuts the front matter from the body: the file opens with a line of three dashes, and the
// next such line closes it.
func split(raw []byte) ([]byte, []byte, error) {
	const fence = "---\n"
	if !bytes.HasPrefix(raw, []byte(fence)) {
		return nil, nil, errors.New("does not open with front matter (---)")
	}
	rest := raw[len(fence):]
	end := bytes.Index(rest, []byte("\n"+fence))
	if end < 0 {
		return nil, nil, errors.New("front matter is not closed (---)")
	}
	return rest[:end+1], rest[end+1+len(fence):], nil
}

var runbookFile = regexp.MustCompile(`^(?:\./)?([a-z0-9]+(?:-[a-z0-9]+)*)\.md(#[a-z0-9-]+)?$`)

// runbookLink is the id a link to another runbook names: its file, beside this one.
func runbookLink(dest string) (string, bool) {
	m := runbookFile.FindStringSubmatch(dest)
	if m == nil {
		return "", false
	}
	return m[1], true
}

// badLink says what is wrong with a link that is not to another runbook: the app opens https and
// mailto, an anchor on the same page, and nothing else.
func badLink(dest string) string {
	switch {
	case strings.HasPrefix(dest, "https://"), strings.HasPrefix(dest, "mailto:"), strings.HasPrefix(dest, "#"):
		return ""
	case strings.HasPrefix(dest, "/"):
		// A path in the app (/admin/jobs): the same installation that serves the runbook.
		return ""
	}
	return fmt.Sprintf("link %q is neither another runbook, a path in the app, https nor mailto", dest)
}
