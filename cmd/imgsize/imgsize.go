package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/nfnt/resize"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"image"
	"image/gif"
	"image/jpeg"
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

var tmpls = template.Must(template.ParseGlob("assets/*.html"))

func handleCreate(rw http.ResponseWriter, req *http.Request) {
	url := req.FormValue("url")
	width, _ := strconv.ParseInt(req.FormValue("width"), 10, 0)
	height, _ := strconv.ParseInt(req.FormValue("height"), 10, 0)
	method := req.FormValue("method")

	hash := md5.New()
	io.WriteString(hash, url)
	binary.Write(hash, binary.LittleEndian, width)
	binary.Write(hash, binary.LittleEndian, height)
	io.WriteString(hash, method)
	hashString := fmt.Sprintf("%x", hash.Sum(nil))

	err := UpsertImage(hashString, url, int(width), int(height), method)
	if err != nil {
		panic(err)
	}

	http.Redirect(rw, req, "/img/"+hashString, http.StatusSeeOther)
}

func handleImg(rw http.ResponseWriter, req *http.Request) {
	img := strings.TrimPrefix(req.URL.Path, "/img/")
	ext := strings.ToLower(path.Ext(img))
	hash := img[:len(img)-len(ext)]

	imgURL, width, height, method, err := SelectImage(hash)
	if err != nil {
		panic(err)
	}

	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif":
	default:
		err = tmpls.ExecuteTemplate(rw, "fmt.html", hash)
		if err != nil {
			panic(err)
		}

		return
	}

	values := make(url.Values, 4)
	values.Add("url", imgURL)
	values.Add("width", strconv.FormatInt(int64(width), 10))
	values.Add("height", strconv.FormatInt(int64(height), 10))
	values.Add("method", method)
	values.Add("fmt", ext)

	newURL := req.URL
	newURL.Path = "/resize"
	newURL.RawQuery = values.Encode()

	req, err = http.NewRequest("get", newURL.String(), nil)
	if err != nil {
		panic(err)
	}

	handleResize(rw, req)
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

	switch q.Get("fmt") {
	case ".png":
		rw.Header().Set("Content-type", "image/png")
		err = png.Encode(rw, img)
	case ".jpg", ".jpeg":
		rw.Header().Set("Content-type", "image/jpeg")
		err = jpeg.Encode(rw, img, nil)
	case ".gif":
		rw.Header().Set("Content-type", "image/gif")
		err = gif.Encode(rw, img, nil)
	}
	if err != nil {
		panic(err)
	}
}

func handleMain(rw http.ResponseWriter, req *http.Request) {
	serve := func(page string) {
		http.ServeFile(rw, req, filepath.Join("assets", page))
	}

	exec := func(page string) {
		err := tmpls.ExecuteTemplate(rw, page, nil)
		if err != nil {
			panic(err)
		}
	}

	switch req.URL.Path {
	case "/favicon.ico":
		serve("favicon.png")
	case "/":
		exec("main.html")
	default:
		serve(req.URL.Path)
	}
}

func main() {
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/img/", handleImg)
	http.HandleFunc("/resize", handleResize)
	http.HandleFunc("/", handleMain)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
