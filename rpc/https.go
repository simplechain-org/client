package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rs/cors"
	"github.com/simplechain-org/client/safe"
	"io"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)


type httpsConn struct {
	client    *http.Client
	req       *http.Request
	closeOnce sync.Once
	closeCh   chan interface{}
}

// httpsConn is treated specially by Client.
func (hc *httpsConn) writeJSON(context.Context, interface{}) error {
	panic("writeJSON called on httpsConn")
}

func (hc *httpsConn) remoteAddr() string {
	return hc.req.URL.String()
}

func (hc *httpsConn) readBatch() ([]*jsonrpcMessage, bool, error) {
	<-hc.closeCh
	return nil, false, io.EOF
}

func (hc *httpsConn) close() {
	hc.closeOnce.Do(func() { close(hc.closeCh) })
}

func (hc *httpsConn) closed() <-chan interface{} {
	return hc.closeCh
}

// HTTPSTimeouts HTTPTimeouts represents the configuration params for the HTTP RPC server.
type HTTPSTimeouts struct {
	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration
}

// DefaultHTTPSTimeouts DefaultHTTPTimeouts represents the default timeout values used if further
// configuration is not provided.
var DefaultHTTPSTimeouts = HTTPSTimeouts{
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
	IdleTimeout:  120 * time.Second,
}

// DialHTTPSWithClient DialHTTPWithClient creates a new RPC client that connects to an RPC server over HTTP
// using the provided HTTP Client.
func DialHTTPSWithClient(endpoint string, client *http.Client) (*Client, error) {
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", contentType)

	initctx := context.Background()
	return newClient(initctx, func(context.Context) (ServerCodec, error) {
		return &httpsConn{client: client, req: req, closeCh: make(chan interface{})}, nil
	})
}

// DialHTTPS creates a new RPC client that connects to an RPC server over HTTP.
func DialHTTPS(endpoint string,certFile string,keyFile string,certFiles []string) (*Client, error) {
	httpClient,err := safe.NewTLSClient(certFile,keyFile,certFiles)
	if err != nil {
		return nil,err
	}
	return DialHTTPSWithClient(endpoint, httpClient)
}

func (c *Client) sendHTTPS(ctx context.Context, op *requestOp, msg interface{}) error {
	hc := c.writeConn.(*httpsConn)
	respBody, err := hc.doRequest(ctx, msg)
	if respBody != nil {
		defer respBody.Close()
	}

	if err != nil {
		if respBody != nil {
			buf := new(bytes.Buffer)
			if _, err2 := buf.ReadFrom(respBody); err2 == nil {
				return fmt.Errorf("%v %v", err, buf.String())
			}
		}
		return err
	}
	var respmsg jsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsg); err != nil {
		return err
	}
	op.resp <- &respmsg
	return nil
}

func (c *Client) sendBatchHTTPS(ctx context.Context, op *requestOp, msgs []*jsonrpcMessage) error {
	hc := c.writeConn.(*httpsConn)
	respBody, err := hc.doRequest(ctx, msgs)
	if err != nil {
		return err
	}
	defer respBody.Close()
	var respmsgs []jsonrpcMessage
	if err := json.NewDecoder(respBody).Decode(&respmsgs); err != nil {
		return err
	}
	for i := 0; i < len(respmsgs); i++ {
		op.resp <- &respmsgs[i]
	}
	return nil
}

func (hc *httpsConn) doRequest(ctx context.Context, msg interface{}) (io.ReadCloser, error) {
	body, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	req := hc.req.WithContext(ctx)
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))

	resp, err := hc.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp.Body, errors.New(resp.Status)
	}
	return resp.Body, nil
}

// httpsServerConn turns a HTTP connection into a Conn.
type httpsServerConn struct {
	io.Reader
	io.Writer
	r *http.Request
}

func newHTTPSServerConn(r *http.Request, w http.ResponseWriter) ServerCodec {
	body := io.LimitReader(r.Body, maxRequestContentLength)
	conn := &httpsServerConn{Reader: body, Writer: w, r: r}
	return NewCodec(conn)
}

// Close does nothing and always returns nil.
func (t *httpsServerConn) Close() error { return nil }

// RemoteAddr returns the peer address of the underlying connection.
func (t *httpsServerConn) RemoteAddr() string {
	return t.r.RemoteAddr
}

// SetWriteDeadline does nothing and always returns nil.
func (t *httpsServerConn) SetWriteDeadline(time.Time) error { return nil }

// NewHTTPSServer reates a new HTTP RPC server around an API provider.
//
// Deprecated: Server implements http.Handler
func NewHTTPSServer(cors []string, vhosts []string, timeouts HTTPSTimeouts, srv http.Handler) *http.Server {
	// Wrap the CORS-handler within a host-handler
	handler := newHttpsCorsHandler(srv, cors)
	handler = newHttpsVHostHandler(vhosts, handler)
	handler = newGzipHandler(handler)

	// Make sure timeout values are meaningful
	if timeouts.ReadTimeout < time.Second {
		fmt.Println("Sanitizing invalid HTTP read timeout", "provided", timeouts.ReadTimeout, "updated", DefaultHTTPSTimeouts.ReadTimeout)
		timeouts.ReadTimeout = DefaultHTTPSTimeouts.ReadTimeout
	}
	if timeouts.WriteTimeout < time.Second {
		fmt.Println("Sanitizing invalid HTTP write timeout", "provided", timeouts.WriteTimeout, "updated", DefaultHTTPSTimeouts.WriteTimeout)
		timeouts.WriteTimeout = DefaultHTTPSTimeouts.WriteTimeout
	}
	if timeouts.IdleTimeout < time.Second {
		fmt.Println("Sanitizing invalid HTTP idle timeout", "provided", timeouts.IdleTimeout, "updated", DefaultHTTPSTimeouts.IdleTimeout)
		timeouts.IdleTimeout = DefaultHTTPSTimeouts.IdleTimeout
	}
	// Bundle and start the HTTP server
	return &http.Server{
		Handler:      handler,
		ReadTimeout:  timeouts.ReadTimeout,
		WriteTimeout: timeouts.WriteTimeout,
		IdleTimeout:  timeouts.IdleTimeout,
	}
}

// ServeHTTPS serves JSON-RPC requests over HTTP.
func (s *Server) ServeHTTPS(w http.ResponseWriter, r *http.Request) {
	// Permit dumb empty requests for remote health-checks (AWS)
	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
		return
	}
	if code, err := validateHttpsRequest(r); err != nil {
		http.Error(w, err.Error(), code)
		return
	}
	// All checks passed, create a codec that reads direct from the request body
	// untilEOF and writes the response to w and order the server to process a
	// single request.
	ctx := r.Context()
	ctx = context.WithValue(ctx, "remote", r.RemoteAddr)
	ctx = context.WithValue(ctx, "scheme", r.Proto)
	ctx = context.WithValue(ctx, "local", r.Host)
	if ua := r.Header.Get("User-Agent"); ua != "" {
		ctx = context.WithValue(ctx, "User-Agent", ua)
	}
	if origin := r.Header.Get("Origin"); origin != "" {
		ctx = context.WithValue(ctx, "Origin", origin)
	}

	w.Header().Set("content-type", contentType)
	codec := newHTTPSServerConn(r, w)
	defer codec.close()
	s.serveSingleRequest(ctx, codec)
}

// validateRequest returns a non-zero response code and error message if the
// request is invalid.
func validateHttpsRequest(r *http.Request) (int, error) {
	if r.Method == http.MethodPut || r.Method == http.MethodDelete {
		return http.StatusMethodNotAllowed, errors.New("method not allowed")
	}
	if r.ContentLength > maxRequestContentLength {
		err := fmt.Errorf("content length too large (%d>%d)", r.ContentLength, maxRequestContentLength)
		return http.StatusRequestEntityTooLarge, err
	}
	// Allow OPTIONS (regardless of content-type)
	if r.Method == http.MethodOptions {
		return 0, nil
	}
	// Check content-type
	if mt, _, err := mime.ParseMediaType(r.Header.Get("content-type")); err == nil {
		for _, accepted := range acceptedContentTypes {
			if accepted == mt {
				return 0, nil
			}
		}
	}
	// Invalid content-type
	err := fmt.Errorf("invalid content type, only %s is supported", contentType)
	return http.StatusUnsupportedMediaType, err
}

func newHttpsCorsHandler(srv http.Handler, allowedOrigins []string) http.Handler {
	// disable CORS support if user has not specified a custom CORS configuration
	if len(allowedOrigins) == 0 {
		return srv
	}
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{http.MethodPost, http.MethodGet},
		MaxAge:         600,
		AllowedHeaders: []string{"*"},
	})
	return c.Handler(srv)
}

// virtualHostHandler is a handler which validates the Domain-header of incoming requests.
// The virtualHostHandler can prevent DNS rebinding attacks, which do not utilize CORS-headers,
// since they do in-domain requests against the RPC api. Instead, we can see on the Domain-header
// which domain was used, and validate that against a whitelist.
type virtualHttpsHostHandler struct {
	vhosts map[string]struct{}
	next   http.Handler
}

// ServeHTTP serves JSON-RPC requests over HTTP, implements http.Handler
func (h *virtualHttpsHostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// if r.Domain is not set, we can continue serving since a browser would set the Domain header
	if r.Host == "" {
		h.next.ServeHTTP(w, r)
		return
	}
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		// Either invalid (too many colons) or no port specified
		host = r.Host
	}
	if ipAddr := net.ParseIP(host); ipAddr != nil {
		// It's an IP address, we can serve that
		h.next.ServeHTTP(w, r)
		return

	}
	// Not an ip address, but a hostname. Need to validate
	if _, exist := h.vhosts["*"]; exist {
		h.next.ServeHTTP(w, r)
		return
	}
	if _, exist := h.vhosts[host]; exist {
		h.next.ServeHTTP(w, r)
		return
	}
	http.Error(w, "invalid host specified", http.StatusForbidden)
}

func newHttpsVHostHandler(vhosts []string, next http.Handler) http.Handler {
	vhostMap := make(map[string]struct{})
	for _, allowedHost := range vhosts {
		vhostMap[strings.ToLower(allowedHost)] = struct{}{}
	}
	return &virtualHttpsHostHandler{vhostMap, next}
}
