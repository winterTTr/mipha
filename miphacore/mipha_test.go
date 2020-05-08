package miphacore

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
)

const (
	testCasesDir      = "testcases"
	templateDir       = "templates"
	configurationName = "configurations.yaml"
	helperName        = "helper.tpl"
	outputDir         = "output"
)

func listFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// skip dir
			if info.IsDir() {
				return nil
			}

			relpath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			files = append(files, relpath)
			return nil
		})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.Compare(files[i], files[j]) < 0
	})

	return files, nil
}

type testCase struct {
	name      string
	useHelper bool
}

func (tc *testCase) Run() error {
	_, filename, _, _ := runtime.Caller(0)
	tdir := filepath.Join(filepath.Dir(filename), testCasesDir)

	tmpdir, err := ioutil.TempDir("", tc.name)
	if err != nil {
		return err
	}

	helper := ""
	if tc.useHelper {
		helper = filepath.Join(tdir, tc.name, helperName)
	}

	m := NewMipha(
		filepath.Join(tdir, tc.name, templateDir),
		filepath.Join(tdir, tc.name, configurationName),
		helper,
		tmpdir)
	defer os.RemoveAll(tmpdir)

	if err := m.Load(); err != nil {
		return err
	}

	if err := m.Execute(); err != nil {
		return err
	}

	// binary compare the whole directory: expected vs generated
	edir := filepath.Join(tdir, tc.name, outputDir)
	efiles, err := listFiles(edir)
	if err != nil {
		return err
	}

	gdir := tmpdir
	gfiles, err := listFiles(gdir)
	if err != nil {
		return err
	}

	if len(efiles) != len(gfiles) {
		return errors.New("output file count not same")
	}

	for i, efile := range efiles {
		gfile := gfiles[i]

		if efile != gfile {
			return fmt.Errorf("expected: %s, got: %s", efile, gfile)
		}

		ec, err := ioutil.ReadFile(filepath.Join(edir, efile))
		if err != nil {
			return err
		}

		gc, err := ioutil.ReadFile(filepath.Join(gdir, gfile))
		if err != nil {
			return err
		}

		if !bytes.Equal(ec, gc) {
			return fmt.Errorf("%s file don't match expectation, content is: %s", gfile, gc)
		}
	}

	return nil
}

func TestSingleTemplate(t *testing.T) {
	tc := testCase{name: "single_template"}
	if err := tc.Run(); err != nil {
		t.Error(err)
	}
}

func TestSingleTemplateUsingHelper(t *testing.T) {
	tc := testCase{name: "single_template_using_helper", useHelper: true}
	if err := tc.Run(); err != nil {
		t.Error(err)
	}
}

func TestSingleTemplateSubFolder(t *testing.T) {
	tc := testCase{name: "single_template_sub_folder"}
	if err := tc.Run(); err != nil {
		t.Error(err)
	}
}
