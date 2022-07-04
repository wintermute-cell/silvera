package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

func testerr(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func compareToCorrect(test_root string, correct_root string, file_path string, t *testing.T) {
	test, err := ioutil.ReadFile(filepath.Join(test_root, file_path))
	testerr(err, t)
	corr, err := ioutil.ReadFile(filepath.Join(correct_root, file_path))
	testerr(err, t)

	if string(test) != string(corr) {
		testerr(fmt.Errorf("%s not equal to reference!", file_path), t)
	}

}

func TestRun(t *testing.T) {
	// get working directory
	TEST_ROOT := filepath.FromSlash("./testdata")
	path := filepath.Join(TEST_ROOT, "test_env")
	WORKING_DIR = path

	old_files, err := filepath.Glob(filepath.Join(WORKING_DIR, "*"))
	testerr(err, t)

	// delete all old files
	for _, f := range old_files {
		if err := os.RemoveAll(f); err != nil {
			t.Error(err)
		}
	}

	err = os.MkdirAll(filepath.Join(TEST_ROOT, "test_results"), 0755)
	testerr(err, t)

	// source path
	SOURCE_DIR = filepath.Join(WORKING_DIR, "src/")
	ADDON_DIR = filepath.Join(WORKING_DIR, "addons/")
	commandInit()

	var checklist []string = []string{ // these files must have been created
		"addons",
		"build",
		"src",
		"silvera.conf",
		"template.html",
	}

	// check if all files from the above list have been created.
	for _, f := range checklist {
		if _, err := os.Stat(filepath.Join(WORKING_DIR, f)); err != nil {
			t.Error(err)
		}
	}

	// overwrite the default config
	defaultConfig := Config{
		Outdir:      filepath.Join(TEST_ROOT, "test_results"),
		Templatedir: filepath.Join(WORKING_DIR, "template.html"),
		Extensions: Exts{
			Table:           true,
			Strikethrough:   true,
			Linkify:         true,
			TaskList:        true,
			DefinitionList:  true,
			Footnote:        true,
			Typographer:     true,
			Wikilink:        true,
			Mathjax:         true,
			TableOfContents: true,
		},
		ParserOptions: ParserOpts{
			WithAttribute:     true,
			WithAutoHeadingID: false,
		},
		RendererOptions: RendererOpts{
			WithHardWraps: false,
			WithXHTML:     true,
			WithUnsafe:    true,
		},
		Addons: []string{},
	}

	// transform the struct to yaml data and write it to a file.
	yamlData, err := yaml.Marshal(&defaultConfig)
	testerr(err, t)
	err = os.WriteFile(filepath.Join(WORKING_DIR, "silvera.conf"), []byte(yamlData), 0644)
	testerr(err, t)

	nestedConfig := Config{
		Outdir:      filepath.Join(TEST_ROOT, "test_results"),
		Templatedir: filepath.Join(WORKING_DIR, "template.html"),
		Extensions: Exts{
			Table:           false,
			Strikethrough:   false,
			Linkify:         false,
			TaskList:        false,
			DefinitionList:  false,
			Footnote:        false,
			Typographer:     false,
			Wikilink:        false,
			Mathjax:         false,
			TableOfContents: false,
		},
		ParserOptions: ParserOpts{
			WithAttribute:     false,
			WithAutoHeadingID: false,
		},
		RendererOptions: RendererOpts{
			WithHardWraps: false,
			WithXHTML:     false,
			WithUnsafe:    false,
		},
		Addons: []string{},
	}

	// transform the struct to yaml data and write it to a file.
	nestedYamlData, err := yaml.Marshal(&nestedConfig)
	testerr(err, t)
	err = os.MkdirAll(filepath.Join(SOURCE_DIR, "subdir1/.slv"), 0755)
	testerr(err, t)
	err = os.WriteFile(filepath.Join(SOURCE_DIR, "subdir1/.slv/silvera.conf"), []byte(nestedYamlData), 0644)
	testerr(err, t)

	// create some test data
	err = os.MkdirAll(filepath.Join(SOURCE_DIR, "subdir1/subdir1subdir1"), 0755)
	testerr(err, t)
	err = os.MkdirAll(filepath.Join(SOURCE_DIR, "subdir1/subdir1subdir2"), 0755)
	testerr(err, t)
	err = os.MkdirAll(filepath.Join(SOURCE_DIR, "subdir2/subdir2subdir1"), 0755)
	testerr(err, t)
	err = os.MkdirAll(filepath.Join(SOURCE_DIR, "subdir2/subdir2subdir2"), 0755)
	testerr(err, t)

	test_markdown := "# Heading\nAnd now some test. In this text we put some $\\frac{1}{2}$ math.\n## Subheading\nAnd now some more: www.some-hyper-link.org. this should trigger the autolink ext.\n \nAfter that lets do a table\n| foo | bar |\n| --- | --- |\n| baz | bim |\n# Heading 2\nHere with some strikethrough:\n~~Hi~~ Hello, world!\n\nAn org style tasklist:\n- [ ] foo\n- [x] bar\n\nA [[wikilink]]\n\n<p>some raw html</p>"
	err = os.WriteFile(filepath.Join(SOURCE_DIR, "subdir1/index.md"), []byte(test_markdown), 0644)
	testerr(err, t)
	err = os.WriteFile(filepath.Join(SOURCE_DIR, "subdir2/index.md"), []byte(test_markdown), 0644)
	testerr(err, t)
	err = os.WriteFile(filepath.Join(SOURCE_DIR, "subdir1/subdir1subdir1/index.md"), []byte(test_markdown), 0644)
	testerr(err, t)
	err = os.WriteFile(filepath.Join(SOURCE_DIR, "subdir2/subdir2subdir1/index.md"), []byte(test_markdown), 0644)
	testerr(err, t)

	commandBuild()

	compareToCorrect(defaultConfig.Outdir, filepath.Join(TEST_ROOT, "correct_results"), "subdir1/index.html", t)
	compareToCorrect(defaultConfig.Outdir, filepath.Join(TEST_ROOT, "correct_results"), "subdir2/index.html", t)
	compareToCorrect(defaultConfig.Outdir, filepath.Join(TEST_ROOT, "correct_results"), "subdir1/subdir1subdir1/index.html", t)
	compareToCorrect(defaultConfig.Outdir, filepath.Join(TEST_ROOT, "correct_results"), "subdir2/subdir2subdir1/index.html", t)
}
