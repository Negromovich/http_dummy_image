package main

import (
	"bytes"
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"image"
	"image/color"
	"image/draw"
	igif "image/gif"
	ijpeg "image/jpeg"
	ipng "image/png"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

var (
	//go:embed font/Roboto-Regular.ttf
	fontBytes []byte
	//go:embed favicon.ico
	faviconBytes   []byte
	faviconModTime = time.Now()
)

type imgFormat string

const (
	png imgFormat = "png"
	jpg imgFormat = "jpg"
	gif imgFormat = "gif"
)

func main() {
	port := flag.Uint("port", 8080, "HTTP port")
	flag.Parse()

	http.HandleFunc("/favicon.ico", faviconHandler)
	http.HandleFunc("/", imageHandler)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		panic(err)
	}
}

func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeContent(w, r, "favicon.ico", faviconModTime, bytes.NewReader(faviconBytes))
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	width, height, text, format, err := parseUrl(r.URL)
	if err != nil {
		statusCode(w, r, 404, err)
		return
	}

	buf := &bytes.Buffer{}
	err = drawImage(buf, width, height, text, format)
	if err != nil {
		statusCode(w, r, 500, err)
		return
	}

	lng, err := w.Write(buf.Bytes())
	if err != nil {
		statusCode(w, r, 500, err)
		return
	}

	switch format {
	case png:
		w.Header().Set("Content-Type", "image/png")
	case jpg:
		w.Header().Set("Content-Type", "image/jpeg")
	case gif:
		w.Header().Set("Content-Type", "image/gif")
	}
	w.Header().Set("Content-Length", strconv.Itoa(lng))
	fmt.Printf("[%s] %d - %s (size %d)\n", time.Now().Format("2006-01-02 15:04:05.000000"), 200, r.URL.Path, lng)
}

func statusCode(rw http.ResponseWriter, req *http.Request, code int, err error) {
	fmt.Printf("[%s] %d - %s - %s\n", time.Now().Format("2006-01-02 15:04:05.000000"), code, req.URL.Path, err)
	if code != 200 {
		rw.WriteHeader(code)
	}
}

func parseUrl(u *url.URL) (width, height int, text string, format imgFormat, err error) {
	exp := regexp.MustCompile(`^(/(?P<text>.*))?/(?P<width>\d+)x(?P<height>\d+)\.(?P<format>png|jpe?g|gif)$`)
	matches := findRegexpMatches(exp, u.Path)

	if val, ok := matches["width"]; !ok {
		err = errors.New("invalid URL, missing width")
		return
	} else {
		width, err = strconv.Atoi(val)
		if err != nil {
			return
		}
	}

	if val, ok := matches["height"]; !ok {
		err = errors.New("invalid URL, missing height")
		return
	} else {
		height, err = strconv.Atoi(val)
		if err != nil {
			return
		}
	}

	if val, ok := matches["format"]; !ok {
		err = errors.New("invalid URL, missing format")
		return
	} else {
		switch val {
		case "png":
			format = png
		case "jpg":
			format = jpg
		case "jpeg":
			format = jpg
		case "gif":
			format = gif
		}
	}

	if val, ok := matches["text"]; !ok || val == "" {
		text = fmt.Sprintf("%dx%d", width, height)
	} else {
		text = val
		switch text {
		case "timestamp":
			text = fmt.Sprintf("%d", time.Now().UnixMilli())
		case "datetime":
			text = time.Now().Format("2006-01-02 15:04:05.000 -0700")
		}
	}

	return
}

func findRegexpMatches(exp *regexp.Regexp, str string) map[string]string {
	match := exp.FindStringSubmatch(str)
	result := make(map[string]string)
	if match == nil {
		return result
	}
	for i, name := range exp.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result
}

func drawImage(w io.Writer, width, height int, text string, format imgFormat) error {
	clr := color.Gray16{Y: 32767}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(clr), image.Point{}, draw.Src)
	err := drawLabel(img, width, height, text)
	if err != nil {
		return fmt.Errorf("failed to draw label: %w", err)
	}
	switch format {
	case png:
		err = ipng.Encode(w, img)
	case jpg:
		err = ijpeg.Encode(w, img, &ijpeg.Options{Quality: 90})
	case gif:
		err = igif.Encode(w, img, &igif.Options{})
	}
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return nil
}

func drawLabel(img *image.RGBA, width, height int, text string) error {
	dc := gg.NewContextForRGBA(img)
	err := loadFontFace(dc, width, height, text)
	if err != nil {
		return fmt.Errorf("failed to load font: %w", err)
	}
	dc.SetRGBA(0, 0, 0, 1)
	dc.DrawStringAnchored(text, float64(width)/2, float64(height)/2, 0.5, 0.5)
	dc.Clip()
	return nil
}

func loadFontFace(dc *gg.Context, width, height int, text string) error {
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return err
	}

	face := truetype.NewFace(f, &truetype.Options{Size: 72})
	dc.SetFontFace(face)

	tw, th := dc.MeasureString(text)
	cw := 0.8 * float64(72*width) / tw
	ch := 0.8 * float64(72*height) / th
	points := math.Floor(math.Min(cw, ch))

	face = truetype.NewFace(f, &truetype.Options{Size: points})
	dc.SetFontFace(face)
	return nil
}
