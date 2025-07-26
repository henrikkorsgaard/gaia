package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matryer/is"
)

// How to test panic: https://stackoverflow.com/a/62028796
func TestStaticFolderPanic(t *testing.T) {

	is := is.New(t)

	defer func() {
		_ = recover()
	}()

	NewServer("path/to/nowhere")

	is.Fail()
}

func TestProxyIntegrationIndex(t *testing.T) {

	is := is.New(t)
	app := httptest.NewServer(NewServer("../static"))
	defer app.Close()
	u := fmt.Sprintf("%v/", app.URL)
	fmt.Println(u)
	r, err := http.Get(u)
	is.NoErr(err)
	body, err := io.ReadAll(r.Body)
	is.NoErr(err)
	is.Equal(strings.Contains(string(body), "<h1>Hello World</h1>"), true)
}
