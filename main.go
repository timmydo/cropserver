package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/oliamb/cutter"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	defaultPort, _ = strconv.Atoi(getEnv("LISTEN_PORT", "8080"))
	file           = flag.String("file", getEnv("IMAGE_FILE", "image.png"), "the path to the file to crop")
	urlpath        = flag.String("url", getEnv("URL_PATH", "/testimage"), "the url path for the crop method")
	port           = flag.Int("port", defaultPort, "port to listen on")
)

var fullImage image.Image
var fullImageType string

func crop(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println(r.Form)
	width, widthErr := strconv.Atoi(r.Form.Get("w"))
	height, heightErr := strconv.Atoi(r.Form.Get("h"))

	if widthErr != nil {
		w.Header().Add("x-message", fmt.Sprintf("Error parsing width: %s", widthErr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if heightErr != nil {
		w.Header().Add("x-message", fmt.Sprintf("Error parsing height: %s", heightErr))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if width < 1 || height < 1 || height > fullImage.Bounds().Max.Y || width > fullImage.Bounds().Max.X {
		w.Header().Add("x-message", fmt.Sprintf("Bad width or height: %d x %d", width, height))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	croppedImg, err := cutter.Crop(fullImage, cutter.Config{
		Width:  width,
		Height: height,
	})

	buffer := new(bytes.Buffer)

	switch fullImageType {
	case "jpeg":
		err = jpeg.Encode(buffer, croppedImg, nil)
	case "png":
		err = png.Encode(buffer, croppedImg)
	case "gif":
		err = gif.Encode(buffer, croppedImg, nil)
	}

	if err == nil {
		// todo: set etag
		w.Header().Set("Cache-Control", "max-age=3600")
		w.Header().Set("Content-Type", "image/"+fullImageType)
		w.Write(buffer.Bytes())
	} else {
		panic(err)
	}
}

func healthz(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(*file); os.IsNotExist(err) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "%s not found", *file)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
	log.Println("/healthz - ok")
}

func main() {
	flag.Parse()
	http.HandleFunc("/", crop)
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Listening on %s\n", addr)

	f, err := os.Open(*file)
	if err != nil {
		log.Fatal("Cannot open file", err)
	}

	fullImage, fullImageType, err = image.Decode(f)

	r := mux.NewRouter()
	r.HandleFunc(*urlpath, crop).Methods("GET")
	r.HandleFunc("/healthz", healthz).Methods("GET")

	srv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
