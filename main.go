package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

func main() {
	port, portPresent := os.LookupEnv("PORT")
	if !portPresent {
		log.Fatal("PORT is not defined, exiting")
		os.Exit(-1)
	}
	log.Print("Starting server on port " + port + " ...")
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/random", snifferHandler)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "OK")
}

func snifferHandler(w http.ResponseWriter, r *http.Request) {

	imageTag, err := getImageFromRemote()
	if err != nil {
		errorResponse(&w, err)
		return
	}

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

func getImageFromRemote() (imageTag string, reason error) {
	req, err := http.NewRequest("GET", os.Getenv("URL"), nil)
	if err != nil {
		return "", err
	}

	ageCookie := http.Cookie{
		Name:  "age-gated",
		Value: os.Getenv("AGE_COOKIE"),
	}
	req.AddCookie(&ageCookie)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return sniffImage(body)
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
