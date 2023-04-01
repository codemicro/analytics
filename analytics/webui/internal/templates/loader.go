package templates

import (
	"embed"
	"github.com/flosch/pongo2/v6"
	"io"
	"path"
)

//go:embed *
var templateFS embed.FS

func TemplateLoader() pongo2.TemplateLoader {
	return &templateLoader{templateFS}
}

type templateLoader struct {
	embed.FS
}

func (tl *templateLoader) Abs(base, name string) string {
	return path.Join(path.Dir(base), name)
}

func (tl *templateLoader) Get(name string) (io.Reader, error) {
	f, err := tl.Open(name)
	if err != nil {
		return nil, err
	}
	return f, nil
}
