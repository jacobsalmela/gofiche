package gofiche

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"time"
)

const (
	gofiche = "gofiche"
)

var (
	s         *GoficheSettings
	t         time.Time
	ts        string
	ct        CounterHandler
	slugChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

type GoficheSettings struct {
	ListenAddr string
	Port       int
	Slug       Slug
	Domain     string
	BufferSize int
	UserName   string
	LogFile    string
	OutDir     string
	Debug      bool
}

type Slug struct {
	Length int
	Slug   string
}

// Each time the URL is visited, a counter is incremented and the value is returned.
type CounterHandler struct {
	counter int
}

func init() {
	// seed for random slug generation
	rand.Seed(time.Now().UnixNano())
	// start the counter at 0
	ct.counter = 0
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	// set timestamps
	t = time.Now()
	ts = t.Format(time.RFC3339)

	ctx := r.Context()
	fmt.Println("============================================================")
	fmt.Printf("[%s] %s - %s %s from %s (%s)...\n", gofiche, ts, r.Method, r.RequestURI, r.RemoteAddr, ctx.Value("serverAddr"))
	if s.Debug {
		dump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Printf("[ERROR] %s - %s\n", ts, err)
		}
		fmt.Printf("[DEBUG] %s - %s\n", ts, dump)
	}
	// add a custom header so it is known responses come from here
	w.Header().Add(fmt.Sprintf("X-%s", gofiche), gofiche)

	// Get the body of the content recieved
	//   Example sending ad-hoc data
	//           curl -X POST -d 'This is the body' http://localhost:9999
	//   Example sending an arbitrary file
	//           curl -X POST --data-binary @file.json http://localhost:9999
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[%s] %s - Could not read body!\n", gofiche, ts)
	}

	// Generate a random slug
	s.Slug.Generate()

	// create the pastes dir and set the output file to outDir/slug.txt
	setupOutdir(s.OutDir)
	sf := filepath.Join(s.OutDir, fmt.Sprintf("%s.txt", s.Slug.Slug))

	// Create the file
	f, err := os.Create(sf)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write the content to the file
	_, err = f.Write(body)
	if err != nil {
		log.Fatal(err)
	}

	// also print the slug path after the data is posted
	io.WriteString(w, fmt.Sprintf("%s:%d/%s\n", s.Domain, s.Port, sf))

	// Print a server-side note of the slug that was writteni
	fmt.Printf("[%s] %s - Wrote %s\n", gofiche, ts, sf)
	fmt.Println("============================================================")
}

func getHelp(w http.ResponseWriter, r *http.Request) {
	var examples = []string{
		fmt.Sprintf("Send me some data:"),
		fmt.Sprintf("curl -X POST --data-binary @foo.txt %s:%d", s.Domain, s.Port),
		fmt.Sprintf("curl -X POST -d 'foobar' %s:%d", s.Domain, s.Port),
		fmt.Sprintf("wget --post-data='foobar string' %s:%d", s.Domain, s.Port),
	}

	for _, e := range examples {
		io.WriteString(w, fmt.Sprintf("%s\n", e))
	}
}

func getCount(w http.ResponseWriter, r *http.Request) {
	ct.Increment()
	io.WriteString(w, fmt.Sprintf("Count, %d!\n", ct.counter))
}

// Increment increases the counter
func (ct *CounterHandler) Increment() {
	fmt.Println(ct.counter)
	ct.counter++
}

// Generates a random slug
func (s *Slug) Generate() {
	cnt := make([]rune, s.Length)
	for i := range cnt {
		cnt[i] = slugChars[rand.Intn(len(slugChars))]
	}
	s.Slug = string(cnt)
}

func setupOutdir(d string) error {
	err := os.MkdirAll(s.OutDir, os.ModePerm)
	if err != nil {
		fmt.Printf("[%s] %s - Could not create paste directory: %s\n", gofiche, ts, s.OutDir)
		return err
	}
	return nil
}

// Serve starts the gofiche server and listens for incoming connections on the specified port
func Serve(settings *GoficheSettings) {
	// map user-defined settings for use througout the codebase
	s = settings

	// Define a new server with defined endpoints
	mux := http.NewServeMux()
	mux.HandleFunc("/", getRoot)
	mux.HandleFunc("/help", getHelp)
	mux.HandleFunc("/count", getCount)

	ctx, cancelCtx := context.WithCancel(context.Background())
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, "serverAddr", l.Addr().Network())
			return ctx
		},
	}

	go func() {
		fmt.Printf("Starting %s on %s...\n", gofiche, server.Addr)
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("server closed\n")
		} else if err != nil {
			fmt.Printf("error listening for server: %s\n", err)
		}
		cancelCtx()
	}()

	<-ctx.Done()
	// Start the server
	// fmt.Printf("Starting %s on %s:%d...\n", gofiche, s.ListenAddr, s.Port)
	// err := http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
