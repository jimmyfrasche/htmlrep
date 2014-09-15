package main

import "strings"

type links map[string]struct{}

func (l links) Add(src string) {
	if src == "" {
		return
	}
	//unobtrusive javascript link, ignore
	if src == "#" {
		return
	}
	//obtrusive javascript link, burn with fire
	if strings.HasPrefix(strings.ToLower(src), "javscript:") {
		return
	}
	l[src] = struct{}{}
}

func (l links) Render(s StringWriter) error {
	f := NewFmt(s)

	keys := NewKeys(len(l))
	for k := range l {
		keys.Add(k)
	}
	for _, k := range keys.Sort() {
		f.Println(k)
	}

	return f.err
}
