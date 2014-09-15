package main

//tag name to list of attrs
type dict map[string]map[string]struct{}

func (d dict) Add(tag string, attrs []string) {
	if d[tag] == nil {
		d[tag] = map[string]struct{}{}
	}
	for _, a := range attrs {
		d[tag][a] = struct{}{}
	}
}

func (d dict) Render(s StringWriter) error {
	f := NewFmt(s)

	keys := NewKeys(len(d))
	for k := range d {
		keys.Add(k)
	}
	for _, k := range keys.Sort() {
		f.Println(k)

		keys := NewKeys(len(d[k]))
		for k := range d[k] {
			keys.Add(k)
		}
		for _, k := range keys.Sort() {
			f.Indentln(k)
		}
	}

	return f.err
}
