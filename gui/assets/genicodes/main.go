package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"strings"
	"text/template"
	"unicode"
)

// Command line options
var (
	oPackage = flag.String("p", "assets", "Package name")
)

const (
	PROGNAME = "genicodes"
	VMAJOR   = 0
	VMINOR   = 1
)

type ConstInfo struct {
	Name  string
	Value string
}

type TemplateData struct {
	Packname string
	Consts   []ConstInfo
}

func main() {

	// Parse command line parameters
	flag.Usage = usage
	flag.Parse()

	// Opens input file
	if len(flag.Args()) == 0 {
		log.Fatal("Input file not supplied")
		return
	}
	finput, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
		return
	}

	// Creates optional output file
	fout := os.Stdout
	if len(flag.Args()) > 1 {
		fout, err = os.Create(flag.Args()[1])
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	defer fout.Close()

	// Parse input file
	var td TemplateData
	td.Packname = *oPackage
	err = parse(finput, &td)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Parses the template
	tmpl := template.New("templ")
	tmpl, err = tmpl.Parse(templText)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Expands template to buffer
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, &td)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Formats buffer as Go source
	p, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
		return
	}

	// Writes formatted source to output file
	fout.Write(p)
}

func parse(fin io.Reader, td *TemplateData) error {

	// Read words from input reader and builds words map
	scanner := bufio.NewScanner(fin)
	for scanner.Scan() {
		// Read next line
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			return err
		}
		// Remove line terminator, spaces and ignore empty lines
		line = strings.Trim(line, "\n ")
		if len(line) == 0 {
			continue
		}

		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}
		name := parts[0]
		code := parts[1]
		nameParts := strings.Split(name, "_")
		for i := 0; i < len(nameParts); i++ {
			nameParts[i] = strings.Title(nameParts[i])
		}
		finalName := strings.Join(nameParts, "")
		// If name starts with number adds prefix
		runes := []rune(finalName)
		if unicode.IsDigit(runes[0]) {
			finalName = "N" + finalName
		}
		td.Consts = append(td.Consts, ConstInfo{Name: finalName, Value: "0x" + code})
	}
	return nil
}

// Shows application usage
func usage() {

	fmt.Fprintf(os.Stderr, "%s v%d.%d\n", PROGNAME, VMAJOR, VMINOR)
	fmt.Fprintf(os.Stderr, "usage: %s [options] <input file> <output file>\n", strings.ToLower(PROGNAME))
	flag.PrintDefaults()
	os.Exit(0)
}

const templText = `//
// This file was generated from the original 'codepoints' file
// from the material design icon fonts:
// https://github.com/google/material-design-icons
//
package {{.Packname}}
const (
	{{range .Consts}}
		{{.Name}} = {{.Value}}
	{{- end}}
)

`
