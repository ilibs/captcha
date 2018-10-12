// Copyright 2011 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"net/http"
	"path"
	"strconv"
	"time"
)

type captchaHandler struct {
	imgWidth  int
	imgHeight int
}

// Server returns a handler that serves HTTP requests with image of captchas. Image dimensions are accepted as
// arguments. The server decides which captcha to serve based on the last URL
// path component: file name part must contain a captcha id, file extension —
// its format (PNG).
//
// For example, for file name "LBm5vMjHDtdUfaWYXiQX.png" it serves an image captcha
// with id "LBm5vMjHDtdUfaWYXiQX"
//
// To reload captcha (get a different solution for the same captcha id), append
// "?reload=x" to URL, where x may be anything (for example, current time or a
// random number to make browsers refetch an image instead of loading it from
// cache).
func Server(imgWidth, imgHeight int) http.Handler {
	return &captchaHandler{imgWidth, imgHeight}
}

func (h *captchaHandler) serve(w http.ResponseWriter, r *http.Request, id, ext string) error {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	var content bytes.Buffer
	switch ext {
	case ".png":
		w.Header().Set("Content-Type", "image/png")
		WriteImage(&content, id, h.imgWidth, h.imgHeight)
	default:
		return ErrNotFound
	}

	http.ServeContent(w, r, id+ext, time.Time{}, bytes.NewReader(content.Bytes()))
	return nil
}

func (h *captchaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, file := path.Split(r.URL.Path)
	ext := path.Ext(file)
	id := file[:len(file)-len(ext)]
	if ext == "" || id == "" {
		http.NotFound(w, r)
		return
	}
	if r.FormValue("reload") != "" {
		Reload(id)
	}
	if r.FormValue("h") != "" {
		ih, err := strconv.Atoi(r.FormValue("h"))
		if err == nil {
			h.imgHeight = ih
		}
	}

	if r.FormValue("w") != "" {
		iw, err := strconv.Atoi(r.FormValue("w"))
		if err == nil {
			h.imgWidth = iw
		}
	}

	if h.serve(w, r, id, ext) == ErrNotFound {
		http.NotFound(w, r)
	}
	// Ignore other errors.
}
