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

// this struct contains the user config values regarding internal goldmark (gm) extensions.
type Exts struct {
	Table          bool `yaml:"tables"`
	Strikethrough  bool `yaml:"strikethrough"`
	Linkify        bool `yaml:"autolinks"`
	TaskList       bool `yaml:"task_list"`
	DefinitionList bool `yaml:"definition_list"`
	Footnote       bool `yaml:"footnotes"`
	Typographer    bool `yaml:"typographer"`
}

// this struct contains the user config values regarding parser options.
type ParserOpts struct {
	WithAttribute     bool `yaml:"custom_heading_attrs"`
	WithAutoHeadingID bool `yaml:"auto_heading_id"`
}

// this struct contains the user config values regarding renderer options.
type RendererOpts struct {
	WithHardWraps bool `yaml:"hard_wraps"`
	WithXHTML     bool `yaml:"xhtml"`
	WithUnsafe    bool `yaml:"unsafe_rendering"`
}

// this struct holds the entire user config, once parsed from the yaml file.
// it is compromised of several structs defined above.
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

//// HELPER FUNCTION
// the following functions fulfill various, non-central functions, and could be called
// an arbitrary amount of times, during arbitrary steps in the build pipeline
// -----------------------------------------------------------------------------------

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

//// FLAG BUILDERS
// the following functions each build a list of 'goldmark' Extenders/Options,
// based on the boolean values found in the config file.
// -------------------------------------------------------------------------

// this function builds a list of 'goldmark' Extenders.
// see https://github.com/yuin/goldmark#built-in-extensions for more
// information on the individual extensions.
func buildExtensionList(conf Config) []gm.Extender {
	var extList []gm.Extender
	e := conf.Extensions
	if e.Table {
		extList = append(extList, gme.Table)
	}
	if e.Strikethrough {
		extList = append(extList, gme.Strikethrough)
	}
	if e.Linkify {
		extList = append(extList, gme.Linkify)
	}
	if e.TaskList {
		extList = append(extList, gme.TaskList)
	}
	if e.DefinitionList {
		extList = append(extList, gme.DefinitionList)
	}
	if e.Footnote {
		extList = append(extList, gme.Footnote)
	}
	if e.Typographer {
		extList = append(extList, gme.Typographer)
	}
	return extList
}

// this function builds a list of 'goldmark' parser-options, based on the boolean
// values found in the config struct that is passed.
// these extend/modify the capabilities of the parsing step.
// see https://github.com/yuin/goldmark#parser-options for more information
func buildParserOptList(conf Config) []gmp.Option {
	po := conf.ParserOptions // ParserOptions is a nested struct in the Config struct
	var parserOptList []gmp.Option
	if po.WithAttribute {
		parserOptList = append(parserOptList, gmp.WithAttribute())
	}
	if po.WithAutoHeadingID {
		parserOptList = append(parserOptList, gmp.WithAutoHeadingID())
	}
	return parserOptList
}

// this function builds a list of 'goldmark' renderer-options, based on the boolean
// values found in the config struct that is passed.
// these extend/modify the capabilities of the rendering step.
// see https://github.com/yuin/goldmark#html-renderer-options for more information
func buildRendererOptList(conf Config) []gmr.Option {
	ro := conf.RendererOptions // RendererOptions is a nested struct in the Config struct
	var rendererOptList []gmr.Option
	if ro.WithHardWraps {
		rendererOptList = append(rendererOptList, gmhtml.WithHardWraps())
	}
	if ro.WithXHTML {
		rendererOptList = append(rendererOptList, gmhtml.WithXHTML())
	}
	if ro.WithUnsafe {
		rendererOptList = append(rendererOptList, gmhtml.WithUnsafe())
	}
	return rendererOptList
}

//// HOOKS
// each of these functions represents the execution of one of the 'hooks'.
// the functions will be called at their respective steps in the build-pipeline
// and thus can be used to run code (or external addons) at these specific points.
// ------------------------------------------------------------------------------

// this hook is part of the 'build' command.
// it is called right at the beginning, before any processing happens.
func hookPre() {
	if PRINT_HOOK {
		fmt.Println("Running pre-hook:")
	}
}

// this hook is part of the 'build' command.
// it is called right before a Markdown file is read for processing.
// this makes it useful for modifying the source (.md) file ahead of processing.
func hookPreFile(filepath string) {
	if PRINT_HOOK {
		fmt.Printf("Running pre-file-hook on %s:\n", filepath)
	}
}

// this hook is part of the 'build' command.
// it is called right before a Markdown file is read for processing.
// this makes it useful for modifying the source (.md) file ahead of processing.
func hookPostFile(filepath string) {
	if PRINT_HOOK {
		fmt.Printf("Running post-file-hook on %s:\n", filepath)
	}
}

// this hook is part of the 'build' command.
// it is called right at the end, after all the processing has finished.
func hookPost() {
	if PRINT_HOOK {
		fmt.Println("Running post-hook:")
	}
}

//// COMMANDS
// the following functions are run whenever their respective commands are
// called by the user. To achieve that, they are put in a dictionary, mapping
// a command string like "build" to a function.
// -------------------------------------------------------------------------

var commands map[string]func()

// here, the string keys are associated to the corresponding functions.
func registerCommands() {
	commands["init"] = commandInit
	commands["build"] = commandBuild
}

// the 'init' command is used to transform an existing directory into a workspace.
// a 'scr' directory and a config file are created, initializing the latter with
// default values defined here.
func commandInit() {
	confpath := filepath.Join(WORKING_DIR, "silvera.conf")
	// check if there is already a config file. if so, assume that this is already a workspace and return.
	if _, err := os.Stat(confpath); err == nil {
		fmt.Println("This directory already appears to be a silvera workspace. Nothing changed.")
		return
	} else {
		// create default config file
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

		// transform the struct to yaml data and write it to a file.
		yamlData, err := yaml.Marshal(&defaultConfig)
		checkerr(err)
		err = os.WriteFile(confpath, []byte(yamlData), 0644)
		checkerr(err)

		// create a src directory
		err = os.MkdirAll(SOURCE_DIR, 0755)
		checkerr(err)

		fmt.Printf("Initialized new silvera workspace at %s\n", WORKING_DIR)
	}
}

// the build command is used to take the contents of the 'src' directory, and build a website
// out of them. Inbetween the steps of this pipeline, various hooks are run (see above for hook definitions).
// Only '.md' files are actually processed and turned into '.html' files, all other files and directories are
// simply copied over.
func commandBuild() {
	config := readConfigFile()
	os.MkdirAll(config.Outdir, 0755) // if necessary, create the build directory as given in the config file.

	hookPre() // run the pre-processing hook

	// recursively walk through the source directory
	filepath.Walk(SOURCE_DIR, func(path string, info os.FileInfo, err error) error {
		// ignore dot- files and directories
		if filepath.Base(path)[0] == '.' || strings.HasPrefix(path, ".") {
			return nil
		}
		// don't stop on errors, just output them. We don't want a single file error to prevent building.
		if err != nil {
			fmt.Println("Err:", err)
			return nil
		}

		relpath := strings.TrimPrefix(path, SOURCE_DIR)  // get the relative path of the source directory.
		outpath := filepath.Join(config.Outdir, relpath) // get the relative path of the build directory.

		// if a directory is encountered, just copy/mirror it over to the build dir.
		if info.IsDir() {
			os.Mkdir(outpath, 0755)
			return nil
			// if a '.md' file is encountered, begin processing it to '.html'
		} else if strings.HasSuffix(relpath, ".md") {
			outpath := strings.TrimSuffix(outpath, ".md")
			outpath = outpath + ".html"
			// run pre-file-processing hook
			hookPreFile(path)
			// process the file to html
			html_bytes, err := renderMdToHtml(path, config)
			if err != nil {
				return err
			}

			fmt.Println("built:", relpath, "->", outpath)

			// write the html byte slice to the file-path determined above, ending in '.md'.
			err = ioutil.WriteFile(outpath, html_bytes, 0644)
			if err != nil {
				// if all went well so far, run the post-file-processing hook
				hookPostFile(outpath)
			}
			return err
			// if some file is encountered that is neither a dir, nor a '.md' file, copy it over to the build
			// directory with no changes made.
		} else {
			srcfile, err := ioutil.ReadFile(path) // read the input file
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(outpath, srcfile, 0644) // write it back to the build dir
			return err
		}
	})

	// run the post-processing hook
	hookPost()
}

//// PROCESSING FUNCTIONS
// the following functions are responsible for converting a given file of one format to another format,
// and then return the converted file as a byte array.
// ---------------------------------------------------------------------------------------------------

// this function takes in a path to a '.md' file, and using the goldmark (gm) package, transforms it to html,
// using the configuration obtained from the users config file to adjust what internal extensions and options to use.
func renderMdToHtml(filepath string, config Config) ([]byte, error) {
	md_bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// register all options and exts as per config file to the processor
	md := gm.New(
		gm.WithExtensions(buildExtensionList(config)...),
		gm.WithParserOptions(buildParserOptList(config)...),
		gm.WithRendererOptions(buildRendererOptList(config)...),
	)
	var buf bytes.Buffer
	err = md.Convert(md_bytes, &buf) // run the converter on the '.md' data that was read above
	html_bytes := buf.Bytes()        // retrieve the result as a slice of bytes
	return html_bytes, err
}

//// MAIN
// the following functions are called directly when running the program, and bootstrap the execution.
// -------------------------------------------------------------------------------------------------

// the init function is run directly before 'main()', and is used to initialize some values.
func init() {
	commands = map[string]func(){}

	// get working directory
	path, err := os.Getwd()
	checkerr(err)
	WORKING_DIR = path

	// source path
	SOURCE_DIR = filepath.Join(WORKING_DIR, "src")
}

// the main function is the entrypoint to this program.
// here, arguments are read, and interpreted as commands.
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
