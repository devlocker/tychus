package tychus

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"
)

type mode uint32

const (
	mode_errored mode = 1 << iota
	mode_serving
)

type proxy struct {
	config   *Configuration
	errorStr string
	mode     mode
	requests chan bool
	revproxy *httputil.ReverseProxy
}

// Returns a newly configured proxy
func newProxy(c *Configuration) *proxy {
	url, err := url.Parse(fmt.Sprintf("%s:%v", "http://localhost", c.AppPort))
	if err != nil {
		c.Logger.Fatal(err)
	}

	revproxy := httputil.NewSingleHostReverseProxy(url)
	revproxy.ErrorLog = log.New(ioutil.Discard, "", 0)

	p := &proxy{
		config:   c,
		revproxy: revproxy,
		mode:     mode_serving,
		requests: make(chan bool),
	}

	return p
}

func (p *proxy) start() error {
	server := &http.Server{Handler: p}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "localhost", p.config.ProxyPort))
	if err != nil {
		return err
	}
	defer listener.Close()

	p.config.Logger.Printf(
		"Proxing requests on port %v to %v",
		strconv.Itoa(p.config.ProxyPort),
		strconv.Itoa(p.config.AppPort),
	)

	p.serve()

	err = server.Serve(listener)
	if err != nil {
		return err
	}

	return nil
}

// Proxy the request to the application server.
func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.requests <- true

	timeout := time.After(time.Second * time.Duration(p.config.Timeout))
	tick := time.Tick(50 * time.Millisecond)

	ctx := r.Context()

	for {
		select {
		case <-tick:
			if p.mode == mode_errored {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(p.errorStr))
				return
			}

			writer := &proxyWriter{res: w}
			p.revproxy.ServeHTTP(writer, r)

			// If the request is "successful" - as in the server responded in
			// some way, return the response to the client.
			if writer.status != http.StatusBadGateway {
				return
			}

		case <-timeout:
			p.config.Logger.Print("Timeout reached")
			w.WriteHeader(http.StatusBadGateway)
			w.Write([]byte("Connection Refused"))

			return

		case <-ctx.Done():
			return
		}
	}
}

func (p *proxy) serve() {
	p.config.Logger.Debug("Proxy: Serving")
	p.mode = mode_serving
}

func (p *proxy) error(err error) {
	p.config.Logger.Debug("Proxy: Error Mode")
	p.mode = mode_errored
	p.errorStr = err.Error()
}

// Wrapper around http.ResponseWriter. Since the proxy works rather naively -
// it just retries requests over and over until it gets a response from the app
// server - we can't use the ResponseWriter that is passed to the handler
// because you cannot call WriteHeader multiple times.
type proxyWriter struct {
	res    http.ResponseWriter
	status int
}

func (w *proxyWriter) WriteHeader(status int) {
	if status == 502 {
		w.status = status
		return
	}

	w.res.WriteHeader(status)
}

func (w *proxyWriter) Write(body []byte) (int, error) {
	return w.res.Write(body)
}

func (w *proxyWriter) Header() http.Header {
	return w.res.Header()
}
