package main

// IMPORTS
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/PuerkitoBio/goquery"
	gmtoc "github.com/abhinav/goldmark-toc"
	gmwiki "github.com/abhinav/goldmark-wikilink"
	gmmathjax "github.com/litao91/goldmark-mathjax"
	gm "github.com/yuin/goldmark"
	gme "github.com/yuin/goldmark/extension"
	gmp "github.com/yuin/goldmark/parser"
	gmr "github.com/yuin/goldmark/renderer"
	gmhtml "github.com/yuin/goldmark/renderer/html"
	"gopkg.in/yaml.v2"
)

//// CONFIG STRUCTS
// the following stucts, combined together, define the `silvera.conf` user config.
// -----------------------------------------------------------------------------------

// this struct contains the user config values regarding internal goldmark (gm) extensions.
type Exts struct {
	Table           bool `yaml:"tables"`
	Strikethrough   bool `yaml:"strikethrough"`
	Linkify         bool `yaml:"autolinks"`
	TaskList        bool `yaml:"task_list"`
	DefinitionList  bool `yaml:"definition_list"`
	Footnote        bool `yaml:"footnotes"`
	Typographer     bool `yaml:"typographer"`
	Wikilink        bool `yaml:"wikilink"`
	Mathjax         bool `yaml:"mathjax"`
	TableOfContents bool `yaml:"table_of_contents"`
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
	Templatedir     string       `yaml:"template"`
	Extensions      Exts         `yaml:"extensions"`
	ParserOptions   ParserOpts   `yaml:"parser_options"`
	RendererOptions RendererOpts `yaml:"renderer_options"`
	Addons          []string     `yaml:"addons"`
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
	ADDON_DIR   string
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

func getFirstHeadingFromHtml(html_content string) string {
	// create a new goquery document. Doing this on a pure string doesn't work, have to use a reader.
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(html_content)))
	checkerr(err)

	heading := doc.Find("h1:first-of-type").Text() // search for the first h1 element

	return heading
}

// reads a config file from the given path, using the parent_conf as a base
func readConfigFile(file_path string, parent_conf Config) Config {
	f, err := ioutil.ReadFile(file_path)
	checkerr(err)

	var conf Config = parent_conf // initialize the new config with its parent. new values will overwrite the old ones.
	err = yaml.Unmarshal(f, &conf)
	checkerr(err)

	fmt.Printf("Read config file at %s\n", file_path)

	return conf
}

// given a file path, will return the config most closely matching that path.
// if no local config exists, return nil.
func getMostSpecificConfig(confMap map[string]Config, file_path string) *Config {
	fileInfo, err := os.Stat(file_path)
	checkerr(err)

	// if the given path points to a file, remove the file from the path and use just the dir path
	var dir_path string
	if fileInfo.IsDir() {
		dir_path = filepath.Clean(file_path)
	} else {
		dir_path = filepath.Dir(file_path)
	}

	for dir_path != SOURCE_DIR { // as long as the path is more specific than the src root, keep searching.
		if conf, ok := confMap[dir_path]; ok {
			fmt.Println("Using local conf for", dir_path)
			return &conf
		} else {
			dir_path = filepath.Clean(strings.TrimSuffix(dir_path, filepath.Base(dir_path))) // shorten the path by its last step, making it less specific
		}
	}
	return nil
}

// this contains the logic for executing all the hooks, since they behave largely the same.
func runHookForPrefix(conf Config, prefix string, args []string) {
	for _, addon_name := range conf.Addons {
		addon_dir := filepath.Join(ADDON_DIR, addon_name)
		files, err := ioutil.ReadDir(addon_dir)
		checkerr(err)
		for _, file := range files {
			if !file.IsDir() && strings.HasPrefix(file.Name(), prefix) {
				cmd := filepath.Join(filepath.Join(ADDON_DIR, addon_name), file.Name())
				var out []byte
				ext := filepath.Ext(file.Name())
				if ext == ".py" {
					args = append([]string{cmd}, args...) // when using an interpreter like python, the program is an arg to the interpreter
					out, err = exec.Command("python", args...).Output()
				} else if ext == ".sh" {
					args = append([]string{cmd}, args...)
					out, err = exec.Command("bash", args...).Output()
				} else {
					out, err = exec.Command(cmd, args...).Output()
				}
				checkerr(err)
				if PRINT_HOOK {
					fmt.Printf("Ran addon: %s\n", cmd)
					fmt.Printf("ADDON_OUT:\n %s\n", out)
				}
			}
		}
	}

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
	// third party extensions
	if e.Wikilink {
		extList = append(extList, &gmwiki.Extender{})
	}
	if e.Mathjax {
		extList = append(extList, gmmathjax.MathJax)
	}
	if e.TableOfContents {
		extList = append(extList, &gmtoc.Extender{})
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
func hookPre(conf Config) {
	runHookForPrefix(conf, "prh__", []string{})
}

// this hook is part of the 'build' command.
// it is called right before a Markdown file is read for processing.
// this makes it useful for modifying the source (.md) file ahead of processing.
func hookPreFile(conf Config, filepath string) {
	runHookForPrefix(conf, "prf__", []string{filepath})
}

// this hook is part of the 'build' command.
// it is called right before a Markdown file is read for processing.
// this makes it useful for modifying the build (.html) file after of processing.
func hookPostFile(conf Config, filepath string) {
	runHookForPrefix(conf, "pof__", []string{filepath})
}

// this hook is part of the 'build' command.
// it is called right at the end, after all the processing has finished.
func hookPost(conf Config) {
	runHookForPrefix(conf, "poh__", []string{})
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
			Outdir:      filepath.Join(WORKING_DIR, "build"),
			Templatedir: filepath.Join(WORKING_DIR, "template.html"),
			Extensions: Exts{
				Table:           true,
				Strikethrough:   true,
				Linkify:         true,
				TaskList:        false,
				DefinitionList:  false,
				Footnote:        false,
				Typographer:     false,
				Wikilink:        true,
				Mathjax:         false,
				TableOfContents: false,
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
			Addons: []string{},
		}

		// transform the struct to yaml data and write it to a file.
		yamlData, err := yaml.Marshal(&defaultConfig)
		checkerr(err)
		err = os.WriteFile(confpath, []byte(yamlData), 0644)
		checkerr(err)

		// create a default template.html file
		err = os.WriteFile(defaultConfig.Templatedir, []byte("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n<title>{{.Title}}</title>\n</head>\n<body>\n{{.Body}}\n</body>\n</html>\n"), 0644)
		checkerr(err)

		// create a src directory
		err = os.MkdirAll(SOURCE_DIR, 0755)
		checkerr(err)

		// create a build directory
		err = os.MkdirAll(defaultConfig.Outdir, 0755)
		checkerr(err)

		// create an addon directory
		err = os.MkdirAll(ADDON_DIR, 0755)
		checkerr(err)

		fmt.Printf("Initialized new silvera workspace at %s\n", WORKING_DIR)
	}
}

// the build command is used to take the contents of the 'src' directory, and build a website
// out of them. Inbetween the steps of this pipeline, various hooks are run (see above for hook definitions).
// Only '.md' files are actually processed and turned into '.html' files, all other files and directories are
// simply copied over.
func commandBuild() {
	config := readConfigFile(filepath.Join(WORKING_DIR, "silvera.conf"), Config{}) // read a new config with an empty parent. This is the global config.
	os.MkdirAll(config.Outdir, 0755)                                               // if necessary, create the build directory as given in the config file.

	hookPre(config) // run the pre-processing hook

	localConfigs := make(map[string]Config) // this map holds configuration structs based on directory names

	// recursively walk through the source directory
	filepath.Walk(SOURCE_DIR, func(path string, info os.FileInfo, err error) error {
		// if a HIDDEN_DIR local config dir is encountered, see if it has a configuration file
		if filepath.Base(path) == HIDDEN_DIR && info.IsDir() {
			conf_path := filepath.Join(path, "silvera.conf") // the path where the config should be located
			if _, err := os.Stat(conf_path); err == nil {    // if the config exists and is readable...
				localConfigs[filepath.Clean(strings.TrimSuffix(path, HIDDEN_DIR))] = readConfigFile(conf_path, config) // ...then load it into the localConfigs map
			}
			return filepath.SkipDir
		}

		// ignore other dot- files and directories
		if filepath.Base(path)[0:1] == "." || strings.HasPrefix(path, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		// don't stop on errors, just output them. We don't want a single file error to prevent building.
		if err != nil {
			fmt.Println("Err:", err)
			return nil
		}

		// determine a configuration used for this path
		var localConf Config
		if c := getMostSpecificConfig(localConfigs, path); c != nil { // if there is a local config, use that one
			localConf = *c
		} else { // if there is no local config, use the global one
			localConf = config
		}

		relpath := strings.TrimPrefix(path, SOURCE_DIR)     // get the relative path in the source directory.
		outpath := filepath.Join(localConf.Outdir, relpath) // get the relative path in the build directory.

		// if a directory is encountered, just copy/mirror it over to the build dir.
		if info.IsDir() {
			os.Mkdir(outpath, 0755)
			return nil
			// if a '.md' file is encountered, begin processing it to '.html'
		} else if strings.HasSuffix(relpath, ".md") {
			outpath := strings.TrimSuffix(outpath, ".md")
			outpath = outpath + ".html"
			// run pre-file-processing hook
			hookPreFile(localConf, path)
			// process the file to html
			html_bytes, err := renderMdToHtml(path, localConf)
			if err != nil {
				return err
			}

			// embed the processed html in the template file
			full_html_bytes := embedHtmlInTemplate(html_bytes, relpath, localConf)

			fmt.Println("built:", relpath, "->", outpath)

			// write the html byte slice to the file-path determined above, ending in '.md'.
			err = ioutil.WriteFile(outpath, full_html_bytes, 0644)
			if err == nil {
				// if all went well so far, run the post-file-processing hook
				hookPostFile(localConf, outpath)
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
			if err == nil {
				fmt.Println("clone:", relpath, "->", outpath)
			}
			return err
		}
	})

	// run the post-processing hook
	hookPost(config)
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

// this function takes in the already processed html as a byte slice, and using golangs html/template
// library, embeds these contents in the template.
func embedHtmlInTemplate(html_contents []byte, curr_path string, config Config) []byte {
	// this struct will hold the data to be embedded into to template
	type EmbeddableContents struct {
		Title string
		Body  string
		Path  string
	}

	tmpl := template.Must(template.ParseFiles(config.Templatedir)) // read in the template file, panicking on failure

	contents := EmbeddableContents{
		Title: getFirstHeadingFromHtml(string(html_contents)),
		Body:  string(html_contents),
		Path:  curr_path,
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, contents)
	total_html := buf.Bytes()

	return total_html
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
	SOURCE_DIR = filepath.Join(WORKING_DIR, "src/")
	ADDON_DIR = filepath.Join(WORKING_DIR, "addons/")
}

// the main function is the entrypoint to this program.
// here, arguments are read, and interpreted as commands.
func main() {
	registerCommands() // register the command functions in the commands map

	// check if enough args are supplied
	if len(os.Args) < 2 {
		fmt.Println("Command missing!")
		printUsage()
		return
	}

	// run command if possible
	cmd := os.Args[1]
	if cmdFunc, ok := commands[cmd]; ok { // if the command we ask for exists...
		cmdFunc()
	} else { // and if it doesn't exist...
		fmt.Printf("Unknown command %s\n", cmd)
		printUsage()
		return
	}
}
