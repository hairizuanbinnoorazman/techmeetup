package app

import (
	"html/template"
	"net/http"
)

type index struct{}

func (i index) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.New("test").Parse(`
<head></head>
<body>
	<h1>Authentication</h1></br>
	<p>This section is to provide a convenient way to do authentication without requiring to fiddle with endpoints</p></br>
	<a href="/auth/meetup/authorize">Meetup Authentication</a></br>
</body>	
`)
	tmpl.Execute(w, nil)
}
