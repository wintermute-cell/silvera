# WIP: MISSING FEATURES
# silvera
> A minimal, markdown based static-site-generator. Extensible in python and bash with readable source-code.

## Motivation
None of the available static-site-generators (ssg) fit my needs.
They are either too big and complex (see [Hugo](https://gohugo.io/)),
or unmaintained (see [zs](https://github.com/zserge/zs)).

I wanted an ssg that met the following requirements:
- **Modifiable**:\
    Everyone that wants to use `silvera` should be able to easily modify its source code.
    Because of that, the code is written relatively verbose, and includes a lot comments,
    to allow even users new to `golang` to make changes.
- **Quickly/easily extensible**:\
    `silvera` already comes with some builtin features that can be enabled through a config file.
    Further extensions is easily possible in whatever language the user wishes. 
    (Called `addons` here, to differentiate between builtin options.)
- **Minimal**:\
    The program should be easily understood, easily installed (contained in a singly binary) and not clutter the users system.
    Providing the user a base for building a very specialized tool is the only goal.
- **Well Documented**:\
    A user should be able to find an explanation for every question regarding `silvera`,
    be it in this document, or in the source-code. No feature should be left unexplained or without example.

## Basic Usage
First, initialize a chosen directory as a `workspace`.
Here you will put all the website-source- and configuration-files.

```bash
mkdir my_new_site
cd my_new_site
silvera init
```

A file called `silvera.conf` will be created, containing some default configuration.
All available configuration fields are found in this file.
There are no further config fields.

And create some content:

```bash
cd src
echo "# Hello World" > touch index.md
mkdir subdirectory
echo "# Hello Subdirectory" > touch subdirectory/index.md
echo "# Other File" > touch subdirectory/other.md
cd ..
```

Now go ahead and create the build directory according to the `outdir` defined in `silvera.conf`,
and then finally, build the files:

```bash
mkdir build
silvera build
```

**Done!** The `outdir` should now look like this:

```
build
|- subdirectory
|    |- index.html
|    |- other.html 
|- index.html
```

## Configuration
Basic configuration is done in two files: `silvera.conf` and `template.html`.

### silvera.conf
- **outdir**: Where to output the final product.
- **extensions**
  - **tables**: [GitHub Flavored Markdown: Tables](https://github.github.com/gfm/#tables-extension-)
  - **strikethrough**: [GitHub Flavored Markdown: Strikethrough](https://github.github.com/gfm/#strikethrough-extension-)
  - **autolinks**: [GitHub Flavored Markdown: Autolinks](https://github.github.com/gfm/#autolinks-extension-)
  - **task_list**: [GitHub Flavored Markdown: Task list items](https://github.github.com/gfm/#task-list-items-extension-)
  - **definition_list**: [PHP Markdown Extra: Definition lists](https://michelf.ca/projects/php-markdown/extra/#def-list)
  - **footnotes**: [PHP Markdown Extra: Footnotes](https://michelf.ca/projects/php-markdown/extra/#footnotes)
  - **typographer**: This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/).
  - **wikilink**: Adds support for `[[wiki]]`-style links: [goldmark-wikilink](https://github.com/abhinav/goldmark-wikilink)
  - **mathjax**: Mathjax support for the goldmark markdown parser: [goldmark-mathjax](https://github.com/litao91/goldmark-mathjax)
  - **table_of_contents**: Adds support for generating tables-of-contents for goldmark documents: [goldmark-toc](https://github.com/abhinav/goldmark-toc)
- **parser_options**
  - **custom_heading_attrs**: Allows for custom attributes on headers like `## heading {#id .className attrName=attrValue class="class1 class2"}`.
  - **auto_heading_id**: Automatically generates ids for each heading. This is required for the `table_of_contents` extension to generate links to the listed headings.
- **renderer_options**
  - **hard_wraps**: Renders newlines as `<br>`.
  - **xhtml**: Renders as `XHTML` (just leave this enabled if you've never heard of XHTML).
  - **unsafe_rendering**: Allow for the rendering of potentially dangerous links or raw HTML. **So if you want to use raw HTML mixed with Markdown, turn this on.**

### template.html
This file specifies the HTML environment, in which the converted Markdown content is put it.
The file does not have to be named `template.html`; you can specify the path to this file in `silvera.conf`, including its name.
The template format follows golangs [text/template](https://pkg.go.dev/text/template) format;
In short, this means you can write normal  `html` (or whatever text you'd like), and use markers like `{{.Title}}` to embed generated content.

There are a number of fixed markers available:
- {{.Title}}: An auto detected title of the page (the contents of the first h1 element).
- {{.Body}}:  The actual contents that were generated from Markdown.
- {{.Path}}:  The respective files relative path.

Take the following as a basic example:
```html
<!DOCTYPE html>
<html>
  <head>
    <title>{{.Title}}</title>
  </head>
  <body>
    <p>You are currently at {{.Path}}</p>
    {{.Body}}
  </body>
</html>
```

### Local Cascading Configuration
Within your `src` directory, every path may contain a `.slv` directory with a `silvera.conf` file inside it.
You can use this directory to specify local `silvera.conf` files,
and to store files that should not be copied to the build directory, such as `template.html` files.
This local configuration is cascaded through the nested directories like this:

Imagine the following structure
```
src
|- subdirectory
|    |- .slv
|         |- silvera.conf  <-- LOCAL CONF
|    |- file2.md
|    |- file3.md 
|    |- sub_subdirectory
|         |- file4.md
|- file1.md
```

This local configuration file will affect `file2`, `file3` and `file4`, but not `file1`.
`file1` will instead be processed using the global configuration found in the workspace.
`sub_subdirectory` could also contain its own `.slv/silvera.conf` that would only affect `file4`.

## Addons
Addons are (mostly user written) pieces of code/script, that extend the functionality of `silvera` from the outside.
They act on their own, without the context of the `silvera` process, using their own data.
The name `addon` is chosen in favor of `extension` to make them easier to differentiate from the builtin extension options (see [silvera.conf](#silveraconf)).
`Plugin` would also not be fitting, as `addons` mostly work "on-top" or "around" `silvera`, and are not part of the programs main process.

*Note: `Addons` may come with their own dependencies, such as a `.py` script requiring a python installation.*

### Using plugins
You first have to install a plugin to be able to use it.
To do this, simply move a plugins directory to your workspaces' `plugin` directory.

Lets take [silvera-html-format](TODO) as an example.
Clone the plugin into your `plugin` directory like this:
```bash
cd plugins
git clone https://github.com/wintermute-cell/silvera-html-format
```
Done! Running `ls` should reveal a single `silvera-html-format` directory inside your `plugins` directory.

Now you can go ahead and enable the plugin.
This is done through [silvera.conf](#silveraconf).
In the `plugins` field, add the directory name of your newly installed plugin (`silvera-html-format`) to the list:
```yaml
plugins:
  - silvera-html-format
```
(also valid:)
```yaml
plugins: ['silvera-html-format']
```

Done! The plugin will now work as expected.
Note that you can also utilize the [local config sytem](#local-cascading-configuration) to manage directory specific plugins.

### Existing plugins
A managed (probably incomplete) list of plugins can be found here: [plugin list](plugin_list.md)

### Writing your own
To begin writing a plugin, create a directory for your files to live in.
Inside that directory, you can place any number of executable files, which **have** to obey the following naming guidelines:

- `prh__FILENAME.ext` for `pre-hook` files.
- `prf__FILENAME.ext` for `pre-file-hook` files.
- `pof__FILENAME.ext` for `post-file-hook` files.
- `poh__FILENAME.ext` for `post-hook` files.

plugin files use the `hook`-system to know when they have to be called, and with what arguments.
There are 4 available hooks in `silvera`:
- pre-hook: Called right at the beginning, before any processing happens.
- pre-file-hook: Called right before a Markdown file is read for processing.
  As a first argument, plugins for this hook are given the path of the respective file.
  This makes it useful for modifying the source (.md) file ahead of processing.
- post-file-hook: Called right before a Markdown file is read for processing.
  As a first argument, plugins for this hook are given the path of the respective file.
  This makes it useful for modifying the build (.html) file after of processing.
- post-hook: Called right at the end, after all the processing has finished.

All plugin files in one directory compose a single plugin.
This way, you could use `...-file-hook` executables to gather data, about all the files, save that data in a temporary file,
and then use the data in a `post-hook` executable. For example gathering titles and building a nav-menu out of these titles at the end.

For a single hook, executables are called in alphabetical order.
You should prefix the filenames with numbers, so indicate a clear order, like `prf__0FILENAME.ext` and `prf__1FILENAME.ext`.