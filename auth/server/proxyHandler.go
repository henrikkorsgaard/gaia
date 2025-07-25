package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/sessions"
)

func proxyHandler(store *sessions.CookieStore, originServer *url.URL) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[reverse proxy server] received request at: %s\n", time.Now())
		fmt.Println(originServer)
		r.Host = originServer.Host
		r.URL.Host = originServer.Host
		r.URL.Scheme = originServer.Scheme
		r.RequestURI = ""

		// save the response from the origin server
		oResp, err := http.DefaultClient.Do(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, err)
			return
		}

		fmt.Println(oResp.Body)

		// return response to the client
		w.WriteHeader(http.StatusOK)
		io.Copy(w, oResp.Body)
	})
}
