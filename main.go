package main

// IMPORTS
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	gm "github.com/yuin/goldmark"
	gme "github.com/yuin/goldmark/extension"
	gmp "github.com/yuin/goldmark/parser"
	gmr "github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

type Exts struct {
	Table          bool `yaml:"tables"`
	Strikethrough  bool `yaml:"strikethrough"`
	Linkify        bool `yaml:"autolinks"`
	TaskList       bool `yaml:"task_list"`
	DefinitionList bool `yaml:"definition_list"`
	Footnote       bool `yaml:"footnotes"`
	Typographer    bool `yaml:"typographer"`
}

type ParserOpts struct {
	WithAttribute     bool `yaml:"custom_heading_attrs"`
	WithAutoHeadingID bool `yaml:"auto_heading_id"`
}

type RendererOpts struct {
	WithHardWraps bool `yaml:"hard_wraps"`
	WithXHTML     bool `yaml:"xhtml"`
	WithUnsafe    bool `yaml:"unsafe_rendering"`
}

type Config struct {
	Outdir          string       `yaml:"outdir"`
	Extensions      Exts         `yaml:"extensions"`
	ParserOptions   ParserOpts   `yaml:"parser_options"`
	RendererOptions RendererOpts `yaml:"renderer_options"`
}

// GLOBAL CONSTANTS
const (
	HIDDEN_DIR = ".slv"
	PRINT_HOOK = false
)

// GLOBAL VARS
var (
	WORKING_DIR string
	SOURCE_DIR  string
)

// HELPER FUNCTION
func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("silvera init   -  Initialize a new silvera workspace")
	fmt.Println("silvera build  -  Build the files from ./src")
}

func readConfigFile() Config {
	f, err := ioutil.ReadFile("silvera.conf")
	checkerr(err)

	var conf Config
	err = yaml.Unmarshal(f, &conf)
	checkerr(err)

	return conf
}

func buildExtensionList(conf Config) []gm.Extender {
	var el []gm.Extender
	e := conf.Extensions
	if e.Table {
		el = append(el, gme.Table)
	}
	if e.Strikethrough {
		el = append(el, gme.Strikethrough)
	}
	if e.Linkify {
		el = append(el, gme.Linkify)
	}
	if e.TaskList {
		el = append(el, gme.TaskList)
	}
	if e.DefinitionList {
		el = append(el, gme.DefinitionList)
	}
	if e.Footnote {
		el = append(el, gme.Footnote)
	}
	if e.Typographer {
		el = append(el, gme.Typographer)
	}
	return el
}

func buildParserOptList(conf Config) []gmp.Option {
	po := conf.ParserOptions
	var pol []gmp.Option
	if po.WithAttribute {
		pol = append(pol, gmp.WithAttribute())
	}
	if po.WithAutoHeadingID {
		pol = append(pol, gmp.WithAutoHeadingID())
	}
	return pol
}

func buildRendererOptList(conf Config) []gmr.Option {
	ro := conf.RendererOptions
	var rol []gmr.Option
	if ro.WithHardWraps {
		rol = append(rol, gmhtml.WithHardWraps())
	}
	if ro.WithXHTML {
		rol = append(rol, gmhtml.WithXHTML())
	}
	if ro.WithUnsafe {
		rol = append(rol, gmhtml.WithUnsafe())
	}
	return rol
}

// HOOKS
func hookPre() {
	if PRINT_HOOK {
		fmt.Println("Running pre-hook:")
	}
}
func hookPreFile(filepath string) {
	if PRINT_HOOK {
		fmt.Printf("Running pre-file-hook on %s:\n", filepath)
	}
}
func hookPostFile(filepath string) {
	if PRINT_HOOK {
		fmt.Printf("Running post-file-hook on %s:\n", filepath)
	}
}
func hookPost() {
	if PRINT_HOOK {
		fmt.Println("Running post-hook:")
	}
}

// COMMANDS
var commands map[string]func()

func registerCommands() {
	commands["init"] = commandInit
	commands["build"] = commandBuild
}

func commandInit() {
	confpath := filepath.Join(WORKING_DIR, "silvera.conf")
	if _, err := os.Stat(confpath); err == nil {
		// if file exists already...
		fmt.Println("This directory already appears to be a silvera workspace. Nothing changed.")
		return
	} else {
		// create config file
		defaultConfig := Config{
			Outdir: filepath.Join(WORKING_DIR, "build"),
			Extensions: Exts{
				Table:          true,
				Strikethrough:  true,
				Linkify:        true,
				TaskList:       false,
				DefinitionList: false,
				Footnote:       false,
				Typographer:    false,
			},
			ParserOptions: ParserOpts{
				WithAttribute:     true,
				WithAutoHeadingID: false,
			},
			RendererOptions: RendererOpts{
				WithHardWraps: false,
				WithXHTML:     true,
				WithUnsafe:    false,
			},
		}
		yamlData, err := yaml.Marshal(&defaultConfig)
		checkerr(err)
		err = os.WriteFile(confpath, []byte(yamlData), 0644)
		checkerr(err)

		// create src dir
		err = os.MkdirAll(SOURCE_DIR, 0755)
		checkerr(err)

		fmt.Printf("Initialized new silvera workspace at %s\n", WORKING_DIR)
	}
}

func commandBuild() {
	config := readConfigFile()
	os.MkdirAll(config.Outdir, 0755)

	hookPre()

	// recursively walk through the source directory
	filepath.Walk(SOURCE_DIR, func(path string, info os.FileInfo, err error) error {
		// ignore dot- files and directories
		if filepath.Base(path)[0] == '.' || strings.HasPrefix(path, ".") {
			return nil
		}
		// don't stop on errors
		if err != nil {
			fmt.Println("Err:", err)
			return nil
		}

		relpath := strings.TrimPrefix(path, SOURCE_DIR)
		outpath := filepath.Join(config.Outdir, relpath)
		if info.IsDir() {
			// mirror directory structure in out dir
			os.Mkdir(outpath, 0755)
			return nil
		} else if strings.HasSuffix(relpath, ".md") {
			outpath := strings.TrimSuffix(outpath, ".md")
			outpath = outpath + ".html"
			// run file hook
			hookPreFile(path)
			// process file
			html_bytes, err := renderMdToHtml(path, config)
			if err != nil {
				return err
			}
			fmt.Println("built:", relpath, "->", outpath)
			err = ioutil.WriteFile(outpath, html_bytes, 0644)
			if err != nil {
				hookPostFile(outpath)
			}
			return err
		} else {
			// just copy other file types
			srcfile, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(outpath, srcfile, 0644)
			return err
		}
	})

	hookPost()
}

// PROCESSING FUNCTIONS
func renderMdToHtml(filepath string, config Config) ([]byte, error) {
	md_bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	md := gm.New( // register all options and exts as per config file to the processor
		gm.WithExtensions(buildExtensionList(config)...),
		gm.WithParserOptions(buildParserOptList(config)...),
		gm.WithRendererOptions(buildRendererOptList(config)...),
	)
	var buf bytes.Buffer
	err = md.Convert(md_bytes, &buf)
	html_bytes := buf.Bytes()
	return html_bytes, err
}

// MAIN
func init() {
	commands = map[string]func(){}

	// get working directory
	path, err := os.Getwd()
	checkerr(err)
	WORKING_DIR = path

	// source path
	SOURCE_DIR = filepath.Join(WORKING_DIR, "src")

	//
}

func main() {
	registerCommands()

	// check if enough args are supplied
	if len(os.Args) < 2 {
		fmt.Println("Command missing!")
		printUsage()
		return
	}

	// run command if possible
	cmd := os.Args[1]
	if cmdFunc, ok := commands[cmd]; ok {
		cmdFunc()
	} else {
		fmt.Printf("Unknown command %s\n", cmd)
		printUsage()
		return
	}
}
