

package rpc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPSErrorResponseWithDelete(t *testing.T) {
	testHTTPErrorResponse(t, http.MethodDelete, contentType, "", http.StatusMethodNotAllowed)
}

func TestHTTPSErrorResponseWithPut(t *testing.T) {
	testHTTPErrorResponse(t, http.MethodPut, contentType, "", http.StatusMethodNotAllowed)
}

func TestHTTPSErrorResponseWithMaxContentLength(t *testing.T) {
	body := make([]rune, maxRequestContentLength+1)
	testHTTPErrorResponse(t,
		http.MethodPost, contentType, string(body), http.StatusRequestEntityTooLarge)
}

func TestHTTPSErrorResponseWithEmptyContentType(t *testing.T) {
	testHTTPErrorResponse(t, http.MethodPost, "", "", http.StatusUnsupportedMediaType)
}

func TestHTTPSErrorResponseWithValidRequest(t *testing.T) {
	testHTTPErrorResponse(t, http.MethodPost, contentType, "", 0)
}

func testHTTPSErrorResponse(t *testing.T, method, contentType, body string, expected int) {
	request := httptest.NewRequest(method, "http://url.com", strings.NewReader(body))
	request.Header.Set("content-type", contentType)
	if code, _ := validateRequest(request); code != expected {
		t.Fatalf("response code should be %d not %d", expected, code)
	}
}
