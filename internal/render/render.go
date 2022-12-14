package render

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

const (
	layoutsPath  = "./web/templates/**/*.layout.html"
	viewsPath    = "./web/templates/**/*.view.html"
	partialsPath = "./web/templates/**/*.partial.html"
)

var viewsCache map[string]*template.Template
var cache bool

// SetUp ...
func SetUp(serverCache bool) error {
	cache = serverCache

	vc, err := loadViewsCache()
	if err != nil {
		return err
	}
	viewsCache = vc

	return nil
}

// WriteView ...
func WriteView(w http.ResponseWriter, key string, data interface{}) error {
	templateMap, err := deafultViewsCache()
	if err != nil {
		return err
	}

	v, exist := templateMap[key]
	if !exist {
		return fmt.Errorf("template %s does not exist", key)
	}

	buff := new(bytes.Buffer)
	err = v.Execute(buff, data)

	_, err = buff.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}

func deafultViewsCache() (map[string]*template.Template, error) {
	if cache {
		return viewsCache, nil
	}

	vc, err := loadViewsCache()
	if err != nil {
		return viewsCache, err
	}

	viewsCache = vc
	return viewsCache, nil
}

func loadViewsCache() (map[string]*template.Template, error) {
	vc := map[string]*template.Template{}
	views, err := filepath.Glob(viewsPath)
	if err != nil {
		return vc, err
	}

	for _, v := range views {
		file := filepath.Base(v)
		newView, err := template.New(file).Funcs(templateFuncMap()).ParseFiles(v)
		if err != nil {
			return vc, err
		}

		layouts, err := filepath.Glob(layoutsPath)
		if err != nil {
			return vc, err
		}

		partials, err := filepath.Glob(partialsPath)
		if err != nil {
			return vc, err
		}

		if (len(layouts) > 0) && (len(partials) > 0) {
			newView, err = newView.ParseGlob(layoutsPath)
			newView, err = newView.ParseGlob(partialsPath)
			if err != nil {
				return vc, err
			}
		}
		vc[viewKey(v)] = newView
	}

	return vc, nil
}

func viewKey(path string) string {
	dir := strings.Split(filepath.Dir(path), "/")
	action := strings.Split(filepath.Base(path), ".")
	return fmt.Sprintf("%s_%s", dir[len(dir)-1], action[0])
}

func templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatDate": func(date time.Time) string {
			return date.Format("02 Jan 2006")
		},
		"sliceToStr": func(strSlice []string) string {
			str := ""
			for i, v := range strSlice {
				str += v
				if i < (len(strSlice) - 1) {
					str += ", "
				}
			}
			return str
		},
		"firstCharToUpper": func(str string) string {
			newStr := ""
			for i, c := range str {
				if i == 0 {
					newStr += strings.ToUpper(string(c))
				} else {
					newStr += string(c)
				}
			}
			return newStr
		},
	}
}
