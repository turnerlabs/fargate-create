package build

import (
	"bytes"
	"fmt"
	"log"
	"text/tabwriter"
	"text/template"
)

func check(e error) {
	if e != nil {
		log.Fatal("ERROR: ", e)
	}
}

func applyTemplate(textTemplate string, data interface{}) string {
	//create a formatted template
	buf := new(bytes.Buffer)
	w := tabwriter.NewWriter(buf, 0, 0, 2, ' ', tabwriter.DiscardEmptyColumns)
	tmpl, err := template.New("t").Parse(textTemplate)
	check(err)
	fmt.Fprintln(w)

	//execute the template with the data
	err = tmpl.Execute(w, data)
	check(err)
	w.Flush()
	return buf.String()
}
