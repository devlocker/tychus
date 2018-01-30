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

## Usage

Usage is simple, `tychus` and then your command. On a filesystem change that
command will be rerun.

```
// Go
tychus go run main.go

// Rust
tychus cargo run

// Ruby
tychus ruby myapp.rb

// Shell Commands
tychus ls
```

Need to pass flags? Stick the command in quotes

```
tychus "ruby myapp.rb -e development"
```

Complicated command? Stick it in quotes

```
tychus "go build -o my-bin && echo 'Built Binary' && ./my-bin"
```


## Options
Tychus has a few options. In most cases the defaults should be sufficient. See
below for a few examples.

```yaml
  -a, --app-port int         port your application runs on, overwritten by ENV['PORT'] (default 3000)
  -p, --proxy-port int       proxy port (default 4000)
  -w, --watch stringSlice    comma separated list of extensions that will trigger a reload. If not set, will reload on any file change.
  -x, --ignore stringSlice   comma separated list of directories to ignore file changes in. (default [node_modules,log,tmp,vendor])
  -t, --timeout int          timeout for proxied requests (default 10)

  -h, --help                 help for tychus
      --debug                print debug output
      --no-proxy             will not start proxy if set
      --version              version for tychus
```

Note: If you do not specify any file extensions in `--watch`, Tychus will
trigger a reload on any file change, except for files inside directories listed
in `--ignore`

Note: Tychus will not watch any hidden directories (those beginning with `.`).

## Examples

### Sinatra
By default, Sinatra runs on port `4567`. Only want to watch `ruby` and
`erb` files. Default ignore list is sufficient. The following are equivalent.

```
tychus ruby myapp.rb -w .rb,.erb -a 4567
tychus ruby myapp.rb --watch=.rb,.erb --app-port=4567
```

Visit http://localhost:4000 (4000 is the default proxy host) and to view your
app.


### Foreman / Procfile
Similar to the previous example, except this time running inside of
[foreman](https://github.com/ddollar/foreman) (or someother Procfile runner).

```
# Procfile
web: tychus "rackup -p $PORT -s puma" -w rb,erb
```

Note: If you need to pass flags to your command (like `-p` & `-s` in this case),
wrap your entire command in quotes.

We don't need to explicitly add a `-a $PORT` flag, because `tychus` will
automatically pick up the $PORT and automatically set `app-port` for you.


### Kitchen Sink Example
Running a Go program, separate build and run steps, with some logging thrown in,
only watching `.go` files, running a server on port `5000`, running proxy on
`8080`, ignoring just `tmp` and `vendor`, with a timeout of 5 seconds.

```
tychus "echo 'Building...' && go build -o tmp/my-bin && echo 'Built' && ./tmp/my-bin some args -e development" --app-port=5000 --proxy-port=8080 --watch=.go --ignore=tmp,vendor --timeout=5

# Or, using short flags

tychus "echo 'Building...' && go build -o tmp/my-bin && echo 'Built' && ./tmp/my-bin some args -e development" -a 5000 -p 8080 -w .go -x tmp,vendor -t 5
```

## Whats the point of the proxy?
Consider the following situations:

1. Your server takes ~ 5 seconds to start accepting requests.

```ruby
# myapp.rb
sleep 5
require "sinatra"

get "/"
  "Hello World"
end
```

After your application restarts, any requests that get sent to it within 5
seconds will return an error / show you the "Site can't be reached page".

Really puts a damper on the save, alt+tab, refresh workflow.

By going through the proxy, when you hit refresh, your request will wait until
the server is actually ready to accept and send you back a response. So save,
alt+tab to browser hit refresh. Page will wait the 5 seconds until the server is
ready. Then it will forward the request.

2. You're code has a compile step.

While your code is still compiling you alt+tab to the browser and hit refresh...
and you are potentially served old code. Avoid that by going through a proxy.

**Other Proxy Goodies**

**Error messages**

If you make a syntax error, or your program won't build for some reason, the
output will be displayed in the webpage. Handy for the times you can't see you
server (its in another pane / tab / tmux split).
