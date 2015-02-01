package main

import (
	"crypto/md5"
	"encoding/binary"
	"github.com/nfnt/resize"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

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

func handleCreate(rw http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	url := q.Get("url")
	width, _ := strconv.ParseInt(q.Get("width"), 10, 0)
	height, _ := strconv.ParseInt(q.Get("height"), 10, 0)
	method := q.Get("method")

	hash := md5.New()
	io.WriteString(hash, url)
	binary.Write(hash, binary.LittleEndian, width)
	binary.Write(hash, binary.LittleEndian, height)
	io.WriteString(hash, method)
	hashString := string(hash.Sum(nil))

	err := UpsertImage(hashString, url, int(width), int(height), method)
	if err != nil {
		panic(err)
	}

	http.Redirect(rw, req, "/img/"+hashString+".png", http.StatusSeeOther)
}

func handleImg(rw http.ResponseWriter, req *http.Request) {
	img := strings.TrimPrefix(req.URL.Path, "/img/")
	ext := strings.ToLower(path.Ext(img))
	hash := img[:len(img)-(len(ext)-1)]

	imgURL, width, height, method, err := SelectImage(hash)
	if err != nil {
		panic(err)
	}

	var values url.Values
	values.Add("url", imgURL)
	values.Add("width", strconv.FormatInt(int64(width), 10))
	values.Add("height", strconv.FormatInt(int64(height), 10))
	values.Add("method", method)

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

	rw.Header().Set("Content-Type", "image/png")

	err = png.Encode(rw, img)
	if err != nil {
		panic(err)
	}
}

func handleMain(rw http.ResponseWriter, req *http.Request) {
	var page string
	switch req.URL.Path {
	case "/favicon.ico":
		page = "favicon.png"
	case "/":
		page = "main.html"
	default:
		page = req.URL.Path
	}

	http.ServeFile(rw, req, filepath.Join("assets", page))
}

func main() {
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/img", handleImg)
	http.HandleFunc("/resize", handleResize)
	http.HandleFunc("/", handleMain)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		panic(err)
	}
}
