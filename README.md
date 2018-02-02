tychus
========

Tychus is a command line utility for live reloading applications. Tychus serves
your application through a proxy. Anytime the proxy receives an HTTP request, it
will automatically rerun your command if the filesystem has changed.

`tychus` is language agnostic - it can be configured to work with just about
anything: Go, Rust, Ruby, Python, scripts & arbitrary commands.


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

Usage is simple, `tychus` and then your command. That will start a proxy on port
`4000`. When an HTTP request comes in and the filesystem has changed, your
command will be rerun.

```
tychus go run main.go
```

## Options
Tychus has a few options. In most cases the defaults should be sufficient. See
below for a few examples.

```yaml
  -a, --app-port int         port your application runs on, overwritten by ENV['PORT'] (default 3000)
  -p, --proxy-port int       proxy port (default 4000)
  -x, --ignore stringSlice   comma separated list of directories to ignore file changes in. (default [node_modules,log,tmp,vendor])
      --wait                 Wait for command to finish before proxying a request.
  -t, --timeout int          timeout for proxied requests (default 10)


  -h, --help                 help for tychus
      --debug                print debug output
      --version              version for tychus
```

Note: Tychus will not look for file system changes any hidden directories
(those beginning with `.`).

## Examples

**Example: Web Servers**

```
// Go - Hello World Server
$ tychus go run main.go
[tychus] Proxing requests on port 4000 to 3000
[Go App] App Starting

// Make a request
$ curl localhost:4000
Hello World
$ curl localhost:4000
Hello World

// Save a file, next request will restart your webapp
$ curl localhost:4000
[Go App] App Starting
Hello World
```

This can work with any webserver:

```
// Rust
tychus cargo run

// Ruby
tychus ruby myapp.rb
```

Need to pass flags? Stick the command in quotes

```
tychus "ruby myapp.rb -e development"
```

Complicated command? Stick it in quotes

```
tychus "go build -o my-bin && echo 'Built Binary' && ./my-bin"
```

**Example: Scripts + Commands**

Scenario: You have a webserver running on port `3005`, and it serves static
files from the `/public` directory. In the `/docs` folder are some markdown
files. Should they change, you want them rebuilt and placed into the `public`
directory so the server can pick them up.

```
tychus "multimarkdown docs/index.md > public/index.html" --wait --app-port=3005
```

Now, when you make a request to the proxy on `localhost:4000`, `tychus` will
pause the request (that's what the `--wait` flag is for) until `multimarkdown`
finishes. Then it will forward the request to the server on port `3005`.
`multimarkdown` will only be run if the filesystem has changed.


**Advanced Example: Reload Scripts and Webserver**

Like the scenario above, but you also want your server to autoreload as files
change. You can chain `tychus` together, by setting the `app-port` equal to the
`proxy-port` of the previous `tychus`. An example:

The first instance of `tychus` will run a Go webserver that serves assets out of
`public/`.  We only want it to restart when the `app` folder changes, so ignore
`docs` and `public` directories.

```
$ tychus go run main.go --app-port=3000 --proxy-port=4000 --ignore=docs,public

[tychus] Proxing requests on port 4000 to 3000
...
...
```

In order to serve upto date docs, `multimarkdown` needs to be invoked to
transform markdown into servable html. So we start another `tychus` process to
and point its app-port to server's proxy port.

```
$ tychus "multimarkdown docs/index.md > public/index.html" --wait --app-port=4000 --proxy-port=4001
```

Now, there is a proxy running on `4001` pointing at a proxy on `4000` pointing
at a webserver on `3000`. If you save `docs/index.html`, and then make a request
to `localhost:4001`, that will pause the request while `multimarkdown` runs.
Once it is finished, the requests gets forwarded to `localhost:4000`, which in
turn forwards it our websever on `3000`. The request gets sent all the way back,
with the correctly updated html!

Had our server code been modified in the `app/` folder, then after
`multimarkdown` finished, and the request got passed on to `4000`, that would
have also triggered a restart of our websever.

**Other Proxy Goodies**

**Error messages**

If you make a syntax error, or your program won't build for some reason, stderr
will be returned by the proxy. Handy for the times you can't see you server (its
in another pane / tab / tmux split).
