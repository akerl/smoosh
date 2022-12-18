smoosh
=========

[![GitHub Workflow Status](https://img.shields.io/actions/github/workflow/status/akerl/smoosh/build.yml?branch=main)](https://github.com/akerl/smoosh/actions)
[![GitHub release](https://img.shields.io/github/release/akerl/smoosh.svg)](https://github.com/akerl/smoosh/releases)
[![License](https://img.shields.io/github/license/akerl/smoosh)](https://github.com/akerl/smoosh/blob/master/LICENSE)

Smoosh together files from one or more repos into a target directory. Designed primarily for using git repos to back your dotfiles.

## Usage

Smoosh uses a basic config file:

```
root: /Users/akerl
tmpdir: /Users/akerl/.smoosh/cache
sources:
  - url: https://github.com/akerl/dotfiles
    name: public-dotfiles
    ignore:
      - ^README.md$
```

Several fields are optional:

* root will default to the user's homedir
* tmpdir will default to ~/.smoosh/cache
* A source's name will default to the last element of the URL, as delimited by forward slashes
* A source's ignore list is optional

Smoosh ignores `.gitignore`, `.git`, and `.gitmodules` automatically, regardless of what's set in the ignore list.

The config is read by default from `~/.smoosh/config.yml`, and can be overriden by adding `-c /path/to/your/file`.

To update, run `smoosh update`

## Installation

```
go install github.com/akerl/smoosh@latest
```

## License

smoosh is released under the MIT License. See the bundled LICENSE file for details.
