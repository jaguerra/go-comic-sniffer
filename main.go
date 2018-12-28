package main

import (
	"bytes"
	"fmt"
	"github.com/ericchiang/css"
	"golang.org/x/net/html"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func main() {
	log.Print("Starting server...")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/random", snifferHandler)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func snifferHandler(w http.ResponseWriter, r *http.Request) {

	req, err := http.NewRequest("GET", os.Getenv("URL"), nil)
	if err != nil {
		errorResponse(&w, err)
		return
	}

	ageCookie := http.Cookie{
		Name:  "age-gated",
		Value: os.Getenv("AGE_COOKIE"),
	}
	req.AddCookie(&ageCookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errorResponse(&w, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	imageTag, err := sniffImage(body)
	t, err := template.ParseFiles("random.html")
	if err != nil {
		errorResponse(&w, err)
		return
	}

	err = t.ExecuteTemplate(w, "random.html", template.HTML(imageTag))
	if err != nil {
		errorResponse(&w, err)
		return
	}
}

func errorResponse(wPointer *http.ResponseWriter, reason error) {
	w := *wPointer
	w.WriteHeader(500)
	fmt.Fprintf(w, "KO")
	log.Fatal(reason)
}

func sniffImage(body []byte) (string, error) {
	sel, err := css.Compile(".comic-display img.img-responsive")
	if err != nil {
		return "", err
	}
	node, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	nodes := sel.Select(node)
	if nodes != nil {
		w := bytes.NewBufferString("")
		html.Render(w, nodes[0])
		return w.String(), nil
	}

	return "", nil
}
