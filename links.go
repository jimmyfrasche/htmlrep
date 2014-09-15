package main

type links map[string]struct{}

func (l links) Add(src string) {
	if src == "" {
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
