package main

import (
	"github.com/nfnt/resize"
	"net/http"
	"os"
	"strconv"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

var methods = map[string]resize.InterpolationFunction{
	"nearest-neighbor":   resize.NearestNeighbor,
	"bilinear":           resize.Bilinear,
	"bicubic":            resize.Bicubic,
	"mitchell-netravali": resize.MitchellNetravali,
	"lanczos2":           resize.Lanczos2,
	"lanczos3":           resize.Lanczos3,
}

func handleResize(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	url := q.Get("url")
	width, _ := strconv.ParseInt(q.Get("width"), 10, 0)
	height, _ := strconv.ParseInt(q.Get("height"), 10, 0)
	method := q.Get("method")

	if (width < 0) || (height < 0) {
		panic("Width and height must be positive.")
	}

	if _, ok := methods[method]; !ok {
		panic("Unknown method: " + method)
	}

	rsp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()

	img, _, err := image.Decode(rsp.Body)
	if err != nil {
		panic(err)
	}

	img = resize.Resize(uint(width), uint(height), img, methods[method])
	if err != nil {
		panic(err)
	}

	rw.Header().Set("Content-Type", "image/png")

	err = png.Encode(rw, img)
	if err != nil {
		panic(err)
	}
}

func handleMain(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "assets/main.html")
}

func main() {
	http.HandleFunc("/resize", handleResize)
	http.HandleFunc("/", handleMain)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
