package akhttpd

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func handlePathOrFallback(w http.ResponseWriter, filepath, fallbackContent, contentType string) {
	// create a reader from the filepath
	var reader io.ReadCloser
	var err error
	if filepath != "" {
		reader, err = os.OpenFile(filepath, os.O_RDONLY, os.ModePerm)
	}

	// if something went wrong, or we didn't create a reader in the first place
	// we should use the fallback reader instead.
	if reader == nil || err != nil {
		reader = ioutil.NopCloser(strings.NewReader(fallbackContent))
	}

	writeFile(w, reader, contentType)
}

func writeFile(w http.ResponseWriter, reader io.ReadCloser, contentType string) {
	bytes, err := ioutil.ReadAll(reader)
	defer reader.Close()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(bytes)
}

// This is a minified version of:
//
// <!doctype html>
// <html lang="en">
// <title>akhttpd - Authorized Keys HTTP Daemon</title>
// <style>
// body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif; line-height: 1.5; color: #000; background: #fff; }
// a { color: #000; text-decoration: underline; }
// code { background: lightgray; padding: 5px;  }
// code.replace { user-select: all; }
// code.block { margin: 10px; }
// </style>
// <p>
//     This domain serves an instance of <a href="https://github.com/tkw1536/akhttpd/" rel="noreferrer">akhttpd</a>.
//     Akhttpd serves an <code>authorized_keys</code> file for every GitHub user.
// </p>
// <p>
//     To get the keys for a given user, simply append the username to the URL.
//     For example <code class="replace">http://localhost:8080/username</code> will return an <code>authorized_keys</code> file for the user
//     <code id="editor">username</code>.
// </p>
// <p>
//     To install these keys on an ssh server, you could do something like:
// </p>
// <p>
//     <code class="block replace">
//         curl -L localhost:8080/username > .ssh/authorized_keys
//     </code>
// </p>
// <p>
//     For convenience, this service also exposes a script to do this automatically.
//     Using this script will overwrite any existing SSH Keys for your user.
//     You can use it like:
// </p>
// <p>
//     <code class="block replace">
//         curl -L localhost:8080/username.sh | sh
//     </code>
// </p>
// <p>
//     As you can see this domain is not intended to be used by a web browser.
//     Please also refer to <a href="https://github.com/tkw1536/akhttpd/" rel="noreferrer">the GitHub repository</a>.
// </p>
// <script>
//     var updaters = (function(elements){
//         var result = [];
//         var makeUpdater = function(element) {
// 			var originalText = element.innerHTML;
// 			var hostname = location.host;
// 			var host = location.protocol + "//" + hostname;
//             return function(username) {
//                 element.innerHTML = originalText
//                 .replace('username', username)
//                 .replace('http://localhost:8080', host)
//                 .replace('localhost:8080', hostname);
//             }
// 		};
// 		var updater;
//         for (var i = 0; i < elements.length; i++) {
// 			result.push(updater = makeUpdater(elements[i]));
// 			updater("username");
//         }
//         return result;
//     })(document.getElementsByClassName('replace'));

//     (function(editorElement){
//         editorElement.addEventListener('input', function() {
//             for (var i = 0; i < updaters.length; i++) {
//                 updaters[i](editorElement.innerHTML);
//             }
//         });
//         editorElement.contentEditable = true;
//     })(document.getElementById('editor'));
// </script>
var indexHTML = `<!doctype html><html lang="en"><title>akhttpd - Authorized Keys HTTP Daemon</title><style>body{font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif; line-height: 1.5; color: #000; background: #fff;}a{color: #000; text-decoration: underline;}code{background: lightgray; padding: 5px;}code.replace{user-select: all;}code.block{margin: 10px;}</style><p> This domain serves an instance of <a href="https://github.com/tkw1536/akhttpd/" rel="noreferrer">akhttpd</a>. Akhttpd serves an <code>authorized_keys</code> file for every GitHub user. </p><p> To get the keys for a given user, simply append the username to the URL. For example <code class="replace">http://localhost:8080/username</code> will return an <code>authorized_keys</code> file for the user <code id="editor">username</code>. </p><p> To install these keys on an ssh server, you could do something like:</p><p> <code class="block replace"> curl -L localhost:8080/username > .ssh/authorized_keys </code></p><p> For convenience, this service also exposes a script to do this automatically. Using this script will overwrite any existing SSH Keys for your user. You can use it like:</p><p> <code class="block replace"> curl -L localhost:8080/username.sh | sh </code></p><p> As you can see this domain is not intended to be used by a web browser. Please also refer to <a href="https://github.com/tkw1536/akhttpd/" rel="noreferrer">the GitHub repository</a>.</p><script>var updaters=function(e){for(var n,t=[],r=function(e){var n=e.innerHTML,t=location.host,r=location.protocol+"//"+t;return function(o){e.innerHTML=n.replace("username",o).replace("http://localhost:8080",r).replace("localhost:8080",t)}},o=0;o<e.length;o++)t.push(n=r(e[o])),n("username");return t}(document.getElementsByClassName("replace"));!function(e){e.addEventListener("input",function(){for(var n=0;n<updaters.length;n++)updaters[n](e.innerHTML)}),e.contentEditable=!0}(document.getElementById("editor"));</script>`

var robotsTxt = `
User-agent: *
Disallow: /
Allow: /$
Allow: /_/
Allow: /robots.txt$
`

// ServeUnderscore returns an http.Handler that serves the provided filesystem path under the prefix '_'.
func (Handler) ServeUnderscore(path string) http.Handler {
	return http.StripPrefix("/_/", http.FileServer(noIndexFileSystem{http.Dir(path)}))
}

type noIndexFileSystem struct {
	http.FileSystem
}

func (fs noIndexFileSystem) Open(path string) (http.File, error) {
	file, err := fs.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		return nil, err
	}

	if stat.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := fs.FileSystem.Open(index); err != nil {
			file.Close()
			return nil, err
		}
	}

	return file, nil
}
