# go-manager

## Usage
> This cli template shows the date and time in the terminal

go-manager

## Description

```
This is a template CLI application, which can be used as a boilerplate for awesome CLI tools written in Go.
This template prints the date or time to the terminal.
```
## Examples

```bash
cli-template date
cli-template date --format 20060102
cli-template time
cli-template time --live
```

## Flags
|Flag|Usage|
|----|-----|
|`--debug`|enable debug messages|
|`--disable-update-checks`|disables update checks|
|`--raw`|print unstyled raw output (set it if output is written to a file)|

## Commands
|Command|Usage|
|-------|-----|
|`go-manager completion`|Generate the autocompletion script for the specified shell|
|`go-manager date`|Prints the current date.|
|`go-manager help`|Help about any command|
|`go-manager time`|Prints the current time|
# ... completion
`go-manager completion`

## Usage
> Generate the autocompletion script for the specified shell

go-manager completion

## Description

```
Generate the autocompletion script for go-manager for the specified shell.
See each sub-command's help for details on how to use the generated script.

```

## Commands
|Command|Usage|
|-------|-----|
|`go-manager completion bash`|Generate the autocompletion script for bash|
|`go-manager completion fish`|Generate the autocompletion script for fish|
|`go-manager completion powershell`|Generate the autocompletion script for powershell|
|`go-manager completion zsh`|Generate the autocompletion script for zsh|
# ... completion bash
`go-manager completion bash`

## Usage
> Generate the autocompletion script for bash

go-manager completion bash

## Description

```
Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(go-manager completion bash)

To load completions for every new session, execute once:

#### Linux:

	go-manager completion bash > /etc/bash_completion.d/go-manager

#### macOS:

	go-manager completion bash > $(brew --prefix)/etc/bash_completion.d/go-manager

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion fish
`go-manager completion fish`

## Usage
> Generate the autocompletion script for fish

go-manager completion fish

## Description

```
Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	go-manager completion fish | source

To load completions for every new session, execute once:

	go-manager completion fish > ~/.config/fish/completions/go-manager.fish

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion powershell
`go-manager completion powershell`

## Usage
> Generate the autocompletion script for powershell

go-manager completion powershell

## Description

```
Generate the autocompletion script for powershell.

To load completions in your current shell session:

	go-manager completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... completion zsh
`go-manager completion zsh`

## Usage
> Generate the autocompletion script for zsh

go-manager completion zsh

## Description

```
Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(go-manager completion zsh); compdef _go-manager go-manager

To load completions for every new session, execute once:

#### Linux:

	go-manager completion zsh > "${fpath[1]}/_go-manager"

#### macOS:

	go-manager completion zsh > $(brew --prefix)/share/zsh/site-functions/_go-manager

You will need to start a new shell for this setup to take effect.

```

## Flags
|Flag|Usage|
|----|-----|
|`--no-descriptions`|disable completion descriptions|
# ... date
`go-manager date`

## Usage
> Prints the current date.

go-manager date

## Flags
|Flag|Usage|
|----|-----|
|`-f, --format string`|specify a custom date format (default "02 Jan 06")|
# ... help
`go-manager help`

## Usage
> Help about any command

go-manager help [command]

## Description

```
Help provides help for any command in the application.
Simply type go-manager help [path to command] for full details.
```
# ... time
`go-manager time`

## Usage
> Prints the current time

go-manager time

## Description

```
You can print a live clock with the '--live' flag!
```

## Flags
|Flag|Usage|
|----|-----|
|`-l, --live`|live output|


---
> **Documentation automatically generated with [PTerm](https://github.com/pterm/cli-template) on 12 November 2022**
