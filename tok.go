package main

import (
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"code.google.com/p/go.net/html"
)

func in(s string, cv ...string) bool {
	for _, c := range cv {
		if s == c {
			return true
		}
	}
	return false
}

func inv(s string) func(rune) bool {
	rs := []rune(s)
	return func(c rune) bool {
		for _, r := range rs {
			if r == c {
				return true
			}
		}
		return false
	}
}

//SplitSpace tokenizes a string on whitespace.
func SplitSpace(s string) []string {
	return strings.FieldsFunc(strings.TrimSpace(s), unicode.IsSpace)
}

func isAll(f func(rune) bool) func(string) bool {
	return func(s string) bool {
		for _, r := range s {
			if !f(r) {
				return false
			}
		}
		return true
	}
}

var (
	//IsAlpha reports whether a string consists solely of letters.
	IsAlpha = isAll(unicode.IsLetter)
	//IsNum reports whether a string consists solely of numbers.
	IsNum = isAll(unicode.IsNumber)
)

//AllNum reports whether every s in a []string consists only of numbers.
func AllNum(ss []string) bool {
	for _, s := range ss {
		if !IsNum(s) {
			return false
		}
	}
	return true
}

//IsTitled reports wheter s is in Title Case.
func IsTitled(s string) bool {
	for i, r := range s {
		if i == 0 {
			if !(unicode.IsUpper(r) || unicode.IsTitle(r)) {
				return false
			}
		} else {
			if !unicode.IsLower(r) {
				return false
			}
		}
	}
	return true
}

//these will only be used on tokens without any whitespace
var (
	dotinword = regexp.MustCompile(`[^.]+\.[^.]{2,}`)
	ip6       = regexp.MustCompile(`^[0-9a-zA-Z:]{3,}`)
)

const common = "`—-\"'"

var (
	commonPrefixes = inv("(‛“" + common)
	commonSuffixes = inv("’”.!?,;:)" + common)
)

func isurl(s string) string {
	//do some clean up, to reduce false positives
	s = strings.TrimLeftFunc(s, commonPrefixes)
	if len(s) == 0 {
		return ""
	}

	//we could check for :: here before stripping suffixes,
	//but we are ignoring localhost links

	s = strings.TrimRightFunc(s, commonSuffixes)
	if len(s) == 0 {
		return ""
	}

	//doing this after stripping common suffixes could delete pertinent
	//parts of address but will at least alert of possible presence and
	//it's unlikely that a user would enter an otherwise unadorned ipv6
	//address in the middle of a document at any rate
	if strings.Count(s, ":") > 1 && ip6.MatchString(s) {
		return s
	}

	//if there's a dot midword and it's not an acronym or real
	if dotinword.MatchString(s) {
		components := strings.Split(s, ".")
		//try to detect acronyms
		acro := true
		for _, c := range components {
			if utf8.RuneCountInString(c) == 1 || IsAlpha(c) {
				acro = false
				break
			}
		}

		//if numeric make sure ipv4 and not a real
		allnum := true
		for _, c := range components {
			if !IsNum(c) {
				allnum = false
				break
			}
		}

		if !acro || (allnum && len(components) == 4) {
			return s
		}
	}

	//otherwise consider it a url if contains at least one path separator
	//and is not the very common and/or and not a list of cities, a fraction or a date
	if strings.Index(s, "/") >= 0 && strings.ToLower(s) != "and/or" {
		if s == "/" {
			return s
		}

		//if it's multiple title cased strings joined by /, assume
		//that it's a list of cities like "the Raleigh/Durham area"
		components := strings.Split(s, "/")

		allTitle := true
		for _, c := range components {
			if !IsTitled(c) {
				allTitle = false
				break
			}
		}

		//if all numeric and length 2 a fraction, length 3 a date
		ln := len(components)
		fracdate := (ln == 2 || ln == 3) && AllNum(components)

		if allTitle || fracdate {
			return ""
		}
		return s
	}

	return ""
}

func fmtattr(a html.Attribute) (out string) {
	if a.Namespace != "" {
		out = a.Namespace + ":"
	}
	out += strings.ToLower(a.Key)
	return
}

func attrLinks(tag string, a html.Attribute) (links []string) {
	push := func(l string) []string {
		links = append(links, l)
		return links
	}
	link := func() []string {
		return push(a.Val)
	}

	k := strings.ToLower(a.Key)
	//these will likely be urls regardless of context
	if in(k, "href", "src", "srcdoc", "data-src", "data-href") {
		return link()
	}

	//if not in default namespace we don't know what to look for
	if a.Namespace != "" {
		return
	}

	//these are all the html attributes that will definitely contain a link
	switch {
	case tag == "applet" && in(k, "code", "codebase"),
		tag == "command" && k == "icon", //award yourself 10 bonus points for finding this gem
		tag == "object" && k == "data",
		tag == "form" && k == "action",
		tag == "video" && k == "poster",
		in(tag, "input", "button") && k == "formaction":
		return link()
	}

	//this is supposed to be a url, but often isn't
	if in(tag, "blockquote", "del", "ins", "q") && k == "cite" {
		//consider a word with no spaces a url in this context
		if len(SplitSpace(a.Val)) > 0 {
			return
		}
		if u := isurl(a.Val); u != "" {
			return push(u)
		}
	}

	//this is why no one likes the w3c
	if tag == "img" && k == "srcset" {
		for _, dir := range strings.Split(a.Val, ",") {
			if split := SplitSpace(dir); len(split) > 0 {
				push(split[0])
			}
		}
	}
	return
}

func tokenize(r io.Reader) (d dict, l, c links, err error) {
	z := html.NewTokenizer(r)
	d = dict{}
	l = links{}
	c = links{}

	for {
		tt := z.Next()

		if tt == html.ErrorToken {
			err = z.Err()
			if err == io.EOF {
				err = nil
				break
			}
			return nil, nil, nil, err
		}

		t := z.Token()
		if tt == html.TextToken {
			//raw text, scan for urls
			for _, s := range SplitSpace(t.Data) {
				if u := isurl(s); u != "" {
					c.Add(u)
				}
			}
		} else if tt == html.StartTagToken || tt == html.SelfClosingTagToken {
			tag := strings.ToLower(t.Data)
			attrs := make([]string, 0, len(t.Attr))
			for _, a := range t.Attr {
				attrs = append(attrs, fmtattr(a))
				for _, link := range attrLinks(tag, a) {
					l.Add(link)
				}
			}
			d.Add(tag, attrs)
		}
	}

	return
}
