//Command htmlrep prints a report of tags and links used
//in a UTF-8 encoded HTML blob read from stdin.
//
//The document is never parsed, only tokenized, so many documents
//may be concatenated together.
//
//These reports are useful when preparing to migrate legacy data
//into a new web site.
//
//There are three reports: tags and attributes, links in attributes,
//and links in text nodes.
//
//The tags and attributes report lists each tag used in the blob
//on a line followed by all attributes used on all instances of that tag,
//where each attribute is indented by one tab.
//The reports are separated by blank lines.
//
//Both link reports list all unique links, one per line, in the blob.
//The links in attributes reports links found in all attributes known
//to contain links.
//The links on content scans text nodes for things that may be links,
//using a number of heuristics to cull false positives, which,
//while unicode aware, are largely English-centric.
//
//By default all reports are shown, but some may be hidden using the following
//flags:
//	-t	only show tags and attributes report
//	-l	only show the links reports
//	-c	of the links reports, only show links from text nodes
//	-a	of the links reports, only show links from attributes
//
//EXAMPLES
//
//Show all reports
//	cat *.html | htmlrep
//
//Show only tags and attributes reports
//	cat *.html | htmlrep -t
//
//Show only links reports
//	cat *.html | htmlrep -l
//
//Show only probable links from text nodes:
//	cat *.html | htmlrep -l -c
//
//Show only links from attributes:
//	cat *.html | htmlrep -l -a
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
)

var (
	hideLinks     = flag.Bool("t", false, "only show tags")
	hideTags      = flag.Bool("l", false, "only show links")
	hideAttrLinks = flag.Bool("c", false, "only show links in content")
	hideConLinks  = flag.Bool("a", false, "only show links in attributes")
)

func chk(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func nl() {
	n, err := stdout.WriteString("\n")
	chk(err)
	if n != 1 {
		chk(io.ErrShortWrite)
	}
}

var (
	stdin  = bufio.NewReader(os.Stdin)
	stdout = bufio.NewWriter(os.Stdout)
)

func main() {
	defer stdout.Flush()

	flag.Parse()
	log.SetFlags(0)
	if *hideTags && *hideLinks {
		*hideTags = false
		*hideLinks = false
		log.Println("You asked me to only show both.")
	}
	if *hideAttrLinks && *hideConLinks {
		*hideAttrLinks = false
		*hideConLinks = false
		log.Println("You asked me to only show both kinds of links.")
	}

	d, l, c, err := tokenize(stdin)
	chk(err)

	if !*hideTags {
		chk(d.Render(stdout))
	}

	if !*hideTags && !*hideLinks {
		nl()
	}

	if !*hideLinks {
		if !*hideAttrLinks {
			chk(l.Render(stdout))
		}

		if !*hideAttrLinks && !*hideConLinks {
			nl()
		}

		if !*hideConLinks {
			chk(c.Render(stdout))
		}
	}
}
