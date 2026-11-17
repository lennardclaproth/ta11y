package http

import (
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
)

// immutableAssetPrefix is where SvelteKit puts content-hashed build output, which can be
// cached forever because a change produces a new file name.
const immutableAssetPrefix = "/_app/immutable/"

// SiteHandler serves a single-page app built into dir: existing files as they are, and
// index.html for every other path so the client-side router can take over. index.html is
// never cached, so a deploy is picked up on the next navigation.
func SiteHandler(dir string) http.Handler {
	root := os.DirFS(dir)
	files := http.FileServerFS(root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
		if name != "" && name != "index.html" {
			if info, err := fs.Stat(root, name); err == nil && !info.IsDir() {
				if strings.HasPrefix("/"+name, immutableAssetPrefix) {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFileFS(w, r, root, "index.html")
	})
}
