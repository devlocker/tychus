tychus
========

`tychus` is a command line utility to live-reload your application. `tychus`
will watch your filesystem for changes and automatically recompile and restart
code on change.

`tychus` is language agnostic - it can be configured to work with just about
anything: Go, Rust, Ruby, Python, etc.

Should you desire you can use `tychus` as a proxy to your application. It will
pause requests while your application rebuilds & restarts.


## Installation

### Homebrew on macOS

```
brew tap devlocker/tap
brew install tychus
```

### With Go
Assuming you have a working Go environment and `GOPATH/bin` is in your `PATH`

```
go get github.com/devlocker/tychus
```

### Windows
Currently isn't supported :(

## Getting Started
You will need to create a `.tychus.yml` file configuration file. Easiest way is
to generate one is with:

```
$ tychus init
```

Double check your generated `.tychus.yml` config to make sure it knows which
file extensions to watch.

## Usage

Usage is simple, `tychus run` and then your command. On a filesystem change that
command will be rerun.

```
// Go
tychus run go run main.go

// Rust
tychus run cargo run

// Ruby
tychus ruby myapp.rb

// Shell Commands
tychus run ls
```

Need to pass flags? Stick the command in quotes

```
tychus run "ruby myapp.rb -e development"
```

Complicated command? Stick it in quotes

```
tychus run "go build -o my-bin && echo 'Built Binary' && ./my-bin"
```


## Configuration

```yaml
# List of extentions to watch. A change to a file with one of these extensions
# will trigger a fresh of your application.
extensions:
- .go

# List of folders to not watch. Too many watched files / folders can slow things
down, so try and ignore as much as possible.
ignore:
- node_modules
- tmp
- log
- vendor

# If not enabled, proxy will not start.
proxy_enabled: true

# Port proxy runs on.
proxy_port: 4000

# Port your application runs on. NOTE: a PORT environment will take overwrite
# whatever you put here.
app_port: 3000

# In seconds, how long the proxy will attempt to proxy a request until it
# gives up and returns a 502.
timeout: 10
```
