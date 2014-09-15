package main

import "io"

//A StringWriter wraps the common stdlib WriteString method.
type StringWriter interface {
	WriteString(string) (int, error)
}

//A Renderer has a Render method for outputting to a StringWriter.
type Renderer interface {
	Render(StringWriter) error
}

//Fmt is a monad around StringWriter with methods for all the ways of printing we need.
type Fmt struct {
	StringWriter
	err error
}

//NewFmt wraps a StringWriter in a Fmt.
func NewFmt(s StringWriter) *Fmt {
	return &Fmt{s, nil}
}

//Println writes s on its own line.
func (f *Fmt) Println(s string) {
	if f.err != nil {
		return
	}
	s += "\n"
	n, e := f.WriteString(s)
	if e != nil {
		f.err = e
	} else if n != len(s) {
		f.err = io.ErrShortWrite
	}
}

//Indentln writes s on its own line, indented with one tab.
func (f *Fmt) Indentln(s string) {
	f.Println("\t" + s)
}
