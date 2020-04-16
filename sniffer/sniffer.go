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
	"time"

	"github.com/ericchiang/css"
	circuit "github.com/rubyist/circuitbreaker"
	"golang.org/x/net/html"
)

// Sniffer struct
type Sniffer struct {
	httpClient *circuit.HTTPClient
}

// NewSniffer creates and initialises new Sniffer
func NewSniffer() *Sniffer {
	return &Sniffer{
		httpClient: circuit.NewHTTPClient(time.Second*5, 10, nil),
	}
}

type imageState struct {
	imgTag string
	ok     bool
}

// Handler for the sniffer HTTP action
func (s *Sniffer) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

	images := make(chan *imageState)
	var wg sync.WaitGroup

	for i := 0; i < numImages; i++ {
		wg.Add(1)
		go s.asyncGetImage(images, &wg)
	}

	imageTag := ""

	go func() {
		for image := range images {
			if image.ok {
				imageTag += image.imgTag
			}
			wg.Done()
		}
	}()

	wg.Wait()

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

func (s *Sniffer) asyncGetImage(images chan *imageState, wg *sync.WaitGroup) {
	log.Print("Fetch image requested")
	imageTag, err := s.getImageFromRemote()
	if err != nil {
		log.Fatal("Failed to fetch an image")
		log.Fatal(err)
		images <- &imageState{ok: false}
		return
	}
	log.Print("Fetched image")
	images <- &imageState{imgTag: imageTag, ok: true}
}

func (s *Sniffer) getImageFromRemote() (imageTag string, reason error) {
	req, err := http.NewRequest("GET", os.Getenv("URL"), nil)
	if err != nil {
		return "", err
	}

	ageCookie := http.Cookie{
		Name:  "age-gated",
		Value: os.Getenv("AGE_COOKIE"),
	}
	req.AddCookie(&ageCookie)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return s.sniffImage(body)
}

func (s *Sniffer) sniffImage(body []byte) (string, error) {
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
