package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/matryer/is"
)

// This function changes the directory for where the test is run
// This helps with the static file server issue
// See: https://stackoverflow.com/a/60258660
func init() {
	_, filename, _, _ := runtime.Caller(0)
	// The ".." may change depending on you folder structure
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

func TestProxyIntegrationIndex(t *testing.T) {

	is := is.New(t)
	app := httptest.NewServer(NewServer())
	defer app.Close()
	u := fmt.Sprintf("%v/", app.URL)
	fmt.Println(u)
	r, err := http.Get(u)
	is.NoErr(err)
	body, err := io.ReadAll(r.Body)
	is.NoErr(err)
	fmt.Println(string(body))

}
