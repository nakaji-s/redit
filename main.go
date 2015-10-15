package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	template = `
<!doctype html>

<title>%s</title>
<meta charset="utf-8"/>

<link rel="stylesheet" href="/_assets/CodeMirror/lib/codemirror.css">
<link rel="stylesheet" href="/_assets/CodeMirror/addon/foldgutter.css">
<link rel="stylesheet" href="/_assets/CodeMirror/addon/dialog.css">
<link rel="stylesheet" href="/_assets/CodeMirror/theme/monokai.css">
<script src="/_assets/CodeMirror/lib/codemirror.js"></script>
<script src="/_assets/CodeMirror/addon/brackets.js"></script>
<script src="/_assets/CodeMirror/addon/searchcursor.js"></script>
<script src="/_assets/CodeMirror/addon/search.js"></script>
<script src="/_assets/CodeMirror/addon/dialog.js"></script>
<script src="/_assets/CodeMirror/addon/hardwrap.js"></script>
<script src="/_assets/CodeMirror/addon/foldcode.js"></script>
<script src="/_assets/CodeMirror/addon/foldgutter.js"></script>
<script src="/_assets/CodeMirror/addon/brace-fold.js"></script>
<script src="/_assets/CodeMirror/keymap/sublime.js"></script>
<script src="/_assets/jquery-2.1.1.min.js"></script>
<script src="/_assets/jquery.form.js"></script>

<style type=text/css>
	.CodeMirror {
		width: 100%%;
		height: auto;
		border-top: 1px solid #eee;
		border-bottom: 1px solid #eee;
		line-height: 1.3;
		float: left;
	}
	.CodeMirror-linenumbers { padding: 0 8px; }
</style>

<article>
    <script>
 		var editor = CodeMirror(document.body.getElementsByTagName("article")[0], {
          matchBrackets: true,
          indentUnit: 8,
          tabSize: 8,
          indentWithTabs: true,
          mode: "text",
	      lineNumbers: true,
	      autoCloseBrackets: true,
	      showCursorWhenSelecting: true,
	      theme: "monokai",
	      value: "",
	      keyMap: "sublime",
	      foldGutter: true,
	      extraKeys: {"Ctrl-S": function(cm){ $('#saveForm').submit(); }},
	      gutters: ["CodeMirror-linenumbers", "CodeMirror-foldgutter"]
        });

      // prepare the form when the DOM is ready 
      $(document).ready(function() { 
		  $('#loadForm').submit(function() { 
              $(this).ajaxSubmit({success: function(responseText, statusText, xhr, $form) {
			    editor.setValue(responseText)
		      }}); 
              return false; 
          }); 

		  $('#saveForm').submit(function() {
              $(this).ajaxSubmit({beforeSubmit: function(formData, jqForm, options) {
			    formData[0].value = editor.getValue();
			    return true
              }}); 

              return false; 
          }); 

          $('#loadForm').submit()
      }); 
 

    </script>
    
<form id="loadForm" action="/readfile" method="post"> 
    <input type="hidden" name="data" /> 
</form>
<form id="saveForm" action="/writefile" method="post"> 
    <input type="hidden" name="data" /> 
</form>

  </article>
`
)

var (
	addr = flag.String("http", ":8000", "HTTP service address (e.g., ':8000')")
)

func main() {
	flag.Parse()

	if len(os.Args) != 2 {
		fmt.Println("need to set target filePath")
		os.Exit(1)
	}
	filePath := os.Args[1]

	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("file not found")
			os.Exit(1)
		} else {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Path
		if strings.HasPrefix(name, "/_assets/") {
			a, err := Asset(name[1:])
			if err != nil {
				http.Error(w, "404 page not found", 404)
				return
			}

			w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(name)))
			w.Write(a)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(fmt.Sprintf(template, filePath)))
	})

	http.HandleFunc("/readfile", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, string(b))
	})

	http.HandleFunc("/writefile", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		b := []byte(r.Form["data"][0])
		ioutil.WriteFile(filePath, b, 0644)
	})

	server := &http.Server{
		Addr: *addr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.RequestURI())
			http.DefaultServeMux.ServeHTTP(w, r)
		}),
	}

	fmt.Fprintln(os.Stderr, "Lisning at "+*addr)
	log.Fatal(server.ListenAndServe())
}
