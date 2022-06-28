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
    `silvera` already comes with a lot of builtin features that can be enabled through a config file.
    Further extension of the pipeline is as simple as writing a python or bash script.
- **Minimal**:\
    The program should be easily understood, easily installed / contained in a singly binary, not clutter the users system,
    not generate any files that are not necessary.
    Producing simple HTML with minimal user effort is the most important goal.
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
|-subdirectory
|   |-index.html
|   |-other.html 
|-index.html
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
  - **typographer**: This extension substitutes punctuations with typographic entities like [smartypants](https://daringfireball.net/projects/smartypants/)
  - **wikilink**: Adds support for `[[wiki]]`-style links: [goldmark-wikilink](https://github.com/abhinav/goldmark-wikilink)
  - **mathjax**: Mathjax support for the goldmark markdown parser: [goldmark-mathjax](https://github.com/litao91/goldmark-mathjax)
  - **table_of_contents**: Adds support for generating tables-of-contents for goldmark documents: [goldmark-toc](https://github.com/abhinav/goldmark-toc)
- **parser_options**
  - **custom_heading_attrs**:
  - **auto_heading_id**:
- **renderer_options**
  - **hard_wraps**:
  - **xhtml**:
  - **unsafe_rendering**: