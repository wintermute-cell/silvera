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
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

type MdExts struct {
	NoInstraEmph     bool `yaml:"no_intra_emphasis"`
	Tables           bool `yaml:"tables"`
	FencedCode       bool `yaml:"fenced_code"`
	Autolinking      bool `yaml:"autolinking"`
	Strikethrough    bool `yaml:"strikethrough"`
	HardLineBreak    bool `yaml:"hard_line_break"`
	Footnotes        bool `yaml:"footnotes"`
	PandocTitleblock bool `yaml:"pandoc_titleblock"`
	CustomHeaderIds  bool `yaml:"header_ids"`
	DefinitionLists  bool `yaml:"definition_lists"`
}

type HtmlExts struct {
	TOC                 bool `yaml:"table_of_contents"`     // Generate table of contents with links
	SkipHTML            bool `yaml:"skip_html"`             // Skip preformatted HTML blocks
	SkipImages          bool `yaml:"skip_images"`           // Skip embedded images
	SkipLinks           bool `yaml:"skip_links"`            // Skip all links
	Safelink            bool `yaml:"safe_links"`            // Only link to trusted protocols
	NofollowLinks       bool `yaml:"nofollow_links"`        // Only link with rel="nofollow"
	NoreferrerLinks     bool `yaml:"noreferrer_links"`      // Only link with rel="noreferrer"
	NoopenerLinks       bool `yaml:"noopener_links"`        // Only link with rel="noopener"
	HrefTargetBlank     bool `yaml:"blank_target_links"`    // Add a blank target
	CompletePage        bool `yaml:"complete_page"`         // Generate a complete HTML page
	UseXHTML            bool `yaml:"xhtml"`                 // Generate XHTML output instead of HTML
	FootnoteReturnLinks bool `yaml:"footnote_return_links"` // Generate a link at the end of a footnote to return to the source
	Smartypants         bool `yaml:"smartypants"`           // Enable smart punctuation substitutions
}

type Config struct {
	Outdir         string   `yaml:"outdir"`
	MdExtensions   MdExts   `yaml:"md_extensions"`
	HtmlExtensions HtmlExts `yaml:"html_extensions"`
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

func buildHtmlFlags(c Config) {
}

func buildMdFlags(c Config) {
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
			MdExtensions: MdExts{
				NoInstraEmph:     true,
				Tables:           true,
				FencedCode:       true,
				Autolinking:      true,
				Strikethrough:    true,
				HardLineBreak:    false,
				Footnotes:        false,
				PandocTitleblock: false,
				CustomHeaderIds:  true,
				DefinitionLists:  false,
			},
			HtmlExtensions: HtmlExts{
				TOC:                 false,
				SkipHTML:            false,
				SkipImages:          false,
				SkipLinks:           false,
				Safelink:            false,
				NofollowLinks:       false,
				NoreferrerLinks:     false,
				NoopenerLinks:       false,
				HrefTargetBlank:     false,
				CompletePage:        false,
				UseXHTML:            true,
				FootnoteReturnLinks: false,
				Smartypants:         false,
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

		if info.IsDir() {
			// mirror directory structure in out dir
			os.Mkdir(filepath.Join(config.Outdir, path), 0755)
			return nil
		} else {
			// run file hook
			hookPreFile(path)
			// process file
			fmt.Println("building:", path)
			outpath := filepath.Join(config.Outdir, path)
			html_bytes, err := renderMdToHtml(path, config)
			fmt.Println(string(html_bytes))
			if err == nil {
				return err
			}
			err = ioutil.WriteFile(outpath, html_bytes, 0644)
			if err != nil {
				hookPostFile(outpath)
			}
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
	fmt.Println(string(md_bytes))

	md := gm.New(
		gm.WithExtensions(extension.GFM, extension.Footnote),
		gm.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		gm.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	var buf bytes.Buffer
	err = md.Convert(md_bytes, &buf)
	html_bytes := buf.Bytes()
	//html_bytes := bf.Run(md_bytes, bf.WithExtensions(buildMdFlags(config)), bf.WithRenderer(bf.NewHTMLRenderer(bf.HTMLRendererParameters{
	//	Flags: buildHtmlFlags(config),
	//})))
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
