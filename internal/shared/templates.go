package shared

import (
	"html/template"
	"io"
	"net/http"
)

type TemplateService struct {
	contentViewer *template.Template
	result        *template.Template
	onetime       *template.Template
	onetimeReveal *template.Template
	error         *template.Template
	redirect      *template.Template
	index         *template.Template
	partials      *template.Template
}

func NewTemplateService() *TemplateService {
	return &TemplateService{
		contentViewer: template.Must(template.ParseFiles("web/templates/content-viewer.html")),
		result:        template.Must(template.ParseFiles("web/templates/partials/generic-result.html")),
		onetime:       template.Must(template.ParseFiles("web/templates/onetime.html")),
		onetimeReveal: template.Must(template.ParseFiles("web/templates/partials/onetime-revealed.html")),
		error:         template.Must(template.ParseFiles("web/templates/error.html")),
		redirect:      template.Must(template.ParseFiles("web/templates/redirect.html")),
		index:         loadIndexTemplate(),
		partials:      template.Must(template.ParseGlob("web/templates/partials/*.html")),
	}
}

func loadIndexTemplate() *template.Template {
	tmpl := template.Must(template.ParseGlob("web/templates/partials/*.html"))
	return template.Must(tmpl.ParseFiles("web/templates/index.html"))
}

func (ts *TemplateService) RenderContentViewer(w io.Writer, data any) error {
	return ts.contentViewer.Execute(w, data)
}

func (ts *TemplateService) RenderResult(w io.Writer, data any) error {
	return ts.result.Execute(w, data)
}

func (ts *TemplateService) RenderOnetime(w io.Writer, data any) error {
	return ts.onetime.Execute(w, data)
}

func (ts *TemplateService) RenderOnetimeReveal(w io.Writer, data any) error {
	return ts.onetimeReveal.Execute(w, data)
}

func (ts *TemplateService) RenderError(w http.ResponseWriter, message string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]string{
		"Error": message,
	}

	return ts.error.Execute(w, data)
}

func (ts *TemplateService) RenderRedirect(w io.Writer, data any) error {
	return ts.redirect.Execute(w, data)
}

func (ts *TemplateService) RenderIndexTemplate(w io.Writer, name string, data any) error {
	return ts.index.ExecuteTemplate(w, name, data)
}

func (ts *TemplateService) RenderPartialTemplate(w io.Writer, name string, data any) error {
	return ts.partials.ExecuteTemplate(w, name, data)
}
