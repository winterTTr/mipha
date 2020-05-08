package miphacore

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

type templateFile struct {
	fullpath string
	relpath  string
	name     string
	template *template.Template
}

type templateCollection struct {
	helperPath    string
	helpers       []*template.Template
	templatesPath string
	templates     []*templateFile
}

func newTemplateCollection(templates string, helper string) *templateCollection {
	tc := templateCollection{}
	tc.helperPath = helper
	tc.templatesPath = templates
	return &tc
}

func (tc *templateCollection) load() error {
	// log helper template
	if len(tc.helperPath) != 0 {
		const helperTemlatesName = "__helper__.tpl"
		helperBytes, err := ioutil.ReadFile(tc.helperPath)
		if err != nil {
			return err
		}

		helper, err := template.New(helperTemlatesName).Parse(string(helperBytes))
		if err != nil {
			return err
		}

		for _, helperT := range helper.Templates() {
			if helperT.Name() == helperTemlatesName {
				// skip base template
				continue
			}
			tc.helpers = append(tc.helpers, helperT)
		}
	}

	// load template collection
	if err := filepath.Walk(tc.templatesPath, tc.walkFunc()); err != nil {
		return err
	}

	return nil
}

func (tc *templateCollection) walkFunc() filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(tc.templatesPath, filepath.Dir(path))
		if err != nil {
			return err
		}

		// load template
		t, err := template.New(filepath.Base(path)).Funcs(sprig.TxtFuncMap()).ParseFiles(path)
		if err != nil {
			return err
		}
		t.Option("missingkey=error")
		if tc.helpers != nil {
			for _, helperT := range tc.helpers {
				_, err = t.AddParseTree(helperT.Name(), helperT.Tree)
				if err != nil {
					return err
				}
			}
		}

		template := templateFile{
			fullpath: path,
			relpath:  rel,
			name:     info.Name(),
			template: t,
		}

		tc.templates = append(tc.templates, &template)
		return nil
	}
}

type renderContexts struct {
	path     string
	Contexts map[string]interface{} `yaml:"configurations"`
}

func (c *renderContexts) load() error {
	configBytes, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configBytes, &c)
	if err != nil {
		return err
	}
	if c.Contexts == nil {
		return errors.New("Invalid configuration file, no 'configurations' section")
	}

	return nil
}

// MiphaEngine is the core logic to generate template files into directory
type MiphaEngine struct {
	tc         *templateCollection
	rc         *renderContexts
	outputPath string
	logger     *log.Logger
}

// NewMipha create a new MiphaEngine
func NewMipha(
	templates string,
	config string,
	helper string,
	output string) *MiphaEngine {

	return &MiphaEngine{
		tc: newTemplateCollection(templates, helper),
		rc: &renderContexts{
			path: config,
		},
		outputPath: output,
	}
}

// Load will load all the input from disk into engine for further process
func (m *MiphaEngine) Load() error {
	// load template collection
	if err := m.tc.load(); err != nil {
		return err
	}

	// load contexts
	if err := m.rc.load(); err != nil {
		return err
	}

	return nil
}

// Execute engine to generate output to folders
func (m *MiphaEngine) Execute() error {
	if err := os.RemoveAll(m.outputPath); err != nil {
		return err
	}

	// render template
	for cname, context := range m.rc.Contexts {
		// create output folder
		configOutputdir := filepath.Join(m.outputPath, cname)
		if err := os.MkdirAll(configOutputdir, 0775); err != nil {
			return err
		}
		for _, t := range m.tc.templates {
			tdir := filepath.Join(configOutputdir, t.relpath)
			if _, err := os.Stat(tdir); os.IsNotExist(err) {
				if err := os.MkdirAll(tdir, 0775); err != nil {
					return err
				}
			} else if err != nil {
				return err
			}
			if tfile, err := os.Create(filepath.Join(tdir, t.name)); err == nil {
				defer tfile.Close()
				if err := t.template.Execute(tfile, context); err != nil {
					return err
				}
			} else {
				return err
			}
		}
	}

	return nil
}
