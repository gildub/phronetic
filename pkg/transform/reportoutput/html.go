package reportoutput

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gildub/analyze/pkg/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8sapicore "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:generate go run assets_generate.go

// Output reads report stucture, generates html using go templates and writes it to a file
func htmlOutput(report ReportOutput) {
	path := filepath.Join(env.Config().GetString("WorkDir"), htmlFileName)

	f, err := os.Create(path)
	defer f.Close()

	if err != nil {
		panic(errors.Wrap(err, "unable to create html file"))
	}

	htmlTemplate, err := parseTemplates()
	if err != nil {
		panic(errors.Wrap(err, "unable to parse templates"))
	}

	err = htmlTemplate.Execute(f, report)
	if err != nil {
		panic(errors.Wrap(err, "unable to apply parsed template"))
	}
}

func parseTemplates() (*template.Template, error) {
	var fileStringMap = make(map[string]string)

	cssJSFilesPath := []string{
		"css/bootstrap.min.css",
		"css/styles.css",
		"css/patternfly.min.css",
		"js/bootstrap.min.js",
		"js/jquery-3.3.1.slim.min.js",
		"js/popper.min.js",
		"js/custom.js",
	}

	for _, path := range cssJSFilesPath {
		stringFile, err := readAsset(path)
		if err != nil {
			return nil, err
		}
		fileStringMap[path] = stringFile
	}

	helpersTemplateString, err := readAsset("templates/helpers.gohtml")
	if err != nil {
		return nil, err
	}

	htmlTemplate := template.Must(template.New("html").Funcs(template.FuncMap{
		"bootstrapCSS": func() template.CSS {
			return template.CSS(fileStringMap["css/bootstrap.min.css"])
		},
		"stylesCSS": func() template.CSS {
			return template.CSS(fileStringMap["css/styles.css"])
		},
		"patternflyCSS": func() template.CSS {
			return template.CSS(fileStringMap["css/patternfly.min.css"])
		},
		"bootstrapJS": func() template.JS {
			return template.JS(fileStringMap["js/bootstrap.min.js"])
		},
		"jqueryJS": func() template.JS {
			return template.JS(fileStringMap["js/jquery-3.3.1.slim.min.js"])
		},
		"popperJS": func() template.JS {
			return template.JS(fileStringMap["js/popper.min.js"])
		},
		"customJS": func() template.JS {
			return template.JS(fileStringMap["js/custom.js"])
		},
		"formatQuantity": func(q resource.Quantity) string {
			json, _ := json.Marshal(q)
			return string(json)
		},
		"formatDriver": func(d k8sapicore.PersistentVolumeSource) string {
			json, _ := json.Marshal(d)
			return string(json)
		},
		"incrementIndex": func(i int) int {
			return i + 1
		},
		"pvcIndex": func() *int {
			i := new(int)
			*i = 0
			return i
		},
		"incrementPVCIndex": func(i *int) int {
			*i = *i + 1
			return *i
		},
	}).Parse(helpersTemplateString))

	templatePaths := []string{
		"templates/nodes.gohtml",
		"templates/quotas.gohtml",
		"templates/namespaces.gohtml",
		"templates/pvs.gohtml",
		"templates/storageclasses.gohtml",
		"templates/rbac.gohtml",
		"templates/cluster-report.gohtml",
		"templates/component-report.gohtml",
		"templates/main.gohtml",
	}

	for _, path := range templatePaths {
		stringTemplate, err := readAsset(path)
		if err != nil {
			return nil, err
		}
		htmlTemplate = template.Must(htmlTemplate.Parse(stringTemplate))
	}

	logrus.Infof("Report:Added: %s", htmlFileName)

	return htmlTemplate, nil
}

func readAsset(path string) (string, error) {
	file, err := assets.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}
