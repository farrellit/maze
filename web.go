package main

import (
	"embed"
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

//go:embed webui
var staticfs embed.FS

var (
	queryParamNotFound = fmt.Errorf("No such query parameter")
)

type BaseError struct {
	msg     string
	wrapped error
}

func (be *BaseError) Error() string {
	return be.msg
}
func (be *BaseError) Unwrap() error {
	return be.wrapped
}

type ParamOutOfBoundsError struct {
	*BaseError
}

func getQueryInt(r *http.Request, key string) (int, error) {
	if s, ok := r.URL.Query()[key]; !ok {
		return 0, queryParamNotFound
	} else {
		return strconv.Atoi(s[0])
	}
}

func getQueryInt64(r *http.Request, key string) (int64, error) {
	if s, ok := r.URL.Query()[key]; !ok {
		return 0, queryParamNotFound
	} else {
		return strconv.ParseInt(s[0], 10, 64)
	}
}

type MazeRequest struct {
	x, y, scale int
	seed        int64
}

func (mr *MazeRequest) Path() string {
	return fmt.Sprintf("/%dx%d/%d?s=%d", mr.x, mr.y, mr.seed,mr.scale)
}

func (mr *MazeRequest) RenderSVGMaze(w http.ResponseWriter) {
	//log.Printf("%#v Rendering", *mr)
	m := NewMaze(mr.x, mr.y)
	wc := &WalkingCreator{seed: mr.seed}
	wc.Fill(&m.grid, Coord{0, 0}, Coord{m.x - 1, m.y - 1})
	svgd := SVGRenderer{
		dest:  w,
		scale: mr.scale,
	}
	w.Header().Add("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	svgd.Draw(m)
}

func (mr *MazeRequest) SetFromStrings(x, y, scale, seed string) error {
	var nmr MazeRequest = *mr
	if ix, err := strconv.Atoi(x); err != nil {
		return fmt.Errorf("X value invalid: %s could not be parsed as int", x)
	} else {
		nmr.x = ix
	}
	if iy, err := strconv.Atoi(y); err != nil {
		return fmt.Errorf("Y value invalid: %s could not be parsed as int", y)
	} else {
		nmr.y = iy
	}
	if isc, err := strconv.Atoi(scale); err != nil {
		return fmt.Errorf("Scale value invalid: %s could not be parsed as int", scale)
	} else {
		nmr.scale = isc
	}
	if ise, err := strconv.ParseInt(seed, 10, 64); err != nil {
		return fmt.Errorf("Seed value invalid: %s could not be parsed as int64", seed)
	} else {
		nmr.seed = ise
	}
	*mr = nmr
	if err := nmr.Validate(); err != nil {
		return err
	}
	return nil
}

func (mr *MazeRequest) Validate() error {
	var mindim, maxdim = 3, 256
	for _, d := range []struct {
		v int
		n string
	}{
		{mr.x, "x"},
		{mr.y, "y"},
	} {
		if d.v > maxdim || d.v < mindim {
			return &ParamOutOfBoundsError{&BaseError{
				fmt.Sprintf("%s value %d is out of bounds, must be between %d and %d",
					d.n, d.v, mindim, maxdim),
				nil,
			}}
		}
	}
	if mr.scale <= 0 {
			return &ParamOutOfBoundsError{&BaseError{
				fmt.Sprintf("Scale %s is out of bounds; it must be a positive number", mr.scale),
				nil,
			}}
	}
	return nil
}

func main() {
	if os.Getenv("LAMBDA") == "WEB" {
		log.Printf("Starting in AWS Lambda Web Proxy mode")
		lambda.Start(adapter.Proxy)
	} else {
		server := http.Server{
			Handler: mux,
			Addr: "0.0.0.0:1801",
			ReadTimeout: 5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout: 60 * time.Second,
		}
		log.Fatal(server.ListenAndServe())
	}
}

func ServerMux() *http.ServeMux {
	// generate a maze
	var maze_path_re = regexp.MustCompile(`/api/maze/(?P<x>\d+)x(?P<y>\d+)/(?P<seed>\d+)$`)
	mux := http.NewServeMux()
	mux .HandleFunc("/api/maze/", func(w http.ResponseWriter, r *http.Request) {
		match := maze_path_re.FindStringSubmatch(r.URL.Path)
		if match == nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Request path %s did not match the RE2 regular expression %s", r.URL.Path, maze_path_re.String())
			return
		}
		var mr MazeRequest
		var scalestr string = "25"
		if ss, ok := r.URL.Query()["s"]; ok {
			scalestr = ss[len(ss)-1]
		}
		if err := mr.SetFromStrings(match[1], match[2], scalestr, match[3] ); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, err.Error())
			return
		}
		if mr.seed == 0 {
			// if we pass the Creator 0, it will generate its own seed.  But we want a consistent URL, so 
			// we won't allow that.
			rand.Seed(time.Now().UnixNano())
			mr.seed = rand.Int63()
			//log.Printf("Got 0 seed; redirecting to random seed %d", mr.seed)
			http.Redirect(w, r, mr.Path(), http.StatusSeeOther)
			return
		}
		mr.RenderSVGMaze(w)
	})
	mux.Handle("/webui/", http.FileServer(http.FS(staticfs)))
	if os.Getenv("DEV") == "true" {
		http.Handle("/devui/",
			http.StripPrefix("/devui/", http.FileServer(http.Dir("webui/"))),
		)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/webui/", http.StatusSeeOther)
	})
	return mux
}

var mux *http.ServeMux
var adapter *httpadapter.HandlerAdapter
func init() {
	mux = ServerMux()
	if os.Getenv("LAMBDA") == "WEB" {
		adapter = httpadapter.New(mux)
	}
}
