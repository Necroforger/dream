package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

var eventHandlerTmpl = template.Must(template.New("eventHandler").Funcs(template.FuncMap{
	"constName":      constName,
	"isDiscordEvent": isDiscordEvent,
	"privateName":    privateName,
}).Parse(`package dream

import (
	"github.com/Necroforger/discordgo"
)
// -------------------------------------
// Do not edit: Generated by tools/nextevent
// --------------------------------------

{{range .}}
// Next{{.}} returns the next {{.}} event
func (b *Bot) Next{{.}}() *discordgo.{{.}} {
	return <- b.Next{{.}}C()
}
{{end}}

{{range .}}
// Next{{.}}C returns a channel for the next {{.}} event
func (b *Bot) Next{{.}}C() chan *discordgo.{{.}} {
	out := make(chan *discordgo.{{.}})
	b.DG.AddHandlerOnce(func(s *discordgo.Session, e *discordgo.{{.}}) {
		out <- e
	})
	return out
}
{{end}}

	
`))

func main() {
	var buf bytes.Buffer
	dir := filepath.Dir(".")

	fs := token.NewFileSet()
	parsedFile, err := parser.ParseFile(fs, "tools/events.txt", nil, 0)
	if err != nil {
		log.Fatalf("warning: internal error: could not parse events.txt: %s", err)
		return
	}

	names := []string{}
	for object := range parsedFile.Scope.Objects {
		names = append(names, object)
	}

	sort.Strings(names)
	eventHandlerTmpl.Execute(&buf, names)

	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Println("warning: internal error: invalid Go generated:", err)
		src = buf.Bytes()
	}

	err = ioutil.WriteFile(filepath.Join(dir, strings.ToLower("nextevent.go")), src, 0644)
	if err != nil {
		log.Fatal(buf, "writing output: %s", err)
	}
}

var constRegexp = regexp.MustCompile("([a-z])([A-Z])")

func constCase(name string) string {
	return strings.ToUpper(constRegexp.ReplaceAllString(name, "${1}_${2}"))
}

func isDiscordEvent(name string) bool {
	switch {
	case name == "Connect", name == "Disconnect", name == "Event", name == "RateLimit", name == "Interface":
		return false
	default:
		return true
	}
}

func constName(name string) string {
	if !isDiscordEvent(name) {
		return "__" + constCase(name) + "__"
	}

	return constCase(name)
}

func privateName(name string) string {
	return strings.ToLower(string(name[0])) + name[1:]
}
