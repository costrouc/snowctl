package snowflake

import (
	"bytes"
	"encoding/json"
	"text/template"
)

func jsonMarshal(data interface{}) string {
	message, _ := json.Marshal(data)
	return string(message)
}

func templateToQuery(stmtTemplate string, data interface{}) string {
	var stmt bytes.Buffer
	templ := template.New("query")
	templ = templ.Funcs(template.FuncMap{
		"json": jsonMarshal,
	})
	templ = template.Must(templ.Parse(stmtTemplate))
	templ.Execute(&stmt, data)
	return stmt.String()
}
