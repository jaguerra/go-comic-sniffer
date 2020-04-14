package sniffer

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/ericchiang/css"
	"golang.org/x/net/html"
)

// Handler for the sniffer HTTP action
func Handler(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	numImages, err := strconv.Atoi(query.Get("numImages"))
	if err != nil {
		numImages = 1
	}
	if numImages < 1 {
		numImages = 1
	}
	if numImages > 5 {
		numImages = 5
	}

	images := make(chan string, numImages)
	var wg sync.WaitGroup

	for i := 0; i < numImages; i++ {
		wg.Add(1)
		go asyncGetImage(images, &wg)
	}
	wg.Wait()
	close(images)

	imageTag := ""

	for image := range images {
		imageTag += image
	}

	log.Print(imageTag)

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

func asyncGetImage(images chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Print("Fetch image requested")
	imageTag, err := getImageFromRemote()
	if err != nil {
		log.Fatal("Failed to fetch an image")
		log.Fatal(err)
		return
	}
	log.Print("Fetched image")
	images <- imageTag
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

func errorResponse(wPointer *http.ResponseWriter, reason error) {
	w := *wPointer
	w.WriteHeader(500)
	fmt.Fprintf(w, "KO")
	log.Fatal(reason)
}
