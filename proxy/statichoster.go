package proxy

import (
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// StaticFileHoster mimics static file hoster from NPM
func StaticFileHoster(root, rootFile string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" && rootFile != "" {
			path = rootFile
		}

		target := root + path

		if _, err := os.Stat(target); err == nil {
			mediaType := mime.TypeByExtension(filepath.Ext(target))
			contents, err := ioutil.ReadFile(target)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				if mediaType != "" {
					w.Header().Set("Content-Type", mediaType)
				}

				w.Write(contents)
			}

		} else if os.IsNotExist(err) {
			http.NotFound(w, r)
		}
	}
}
