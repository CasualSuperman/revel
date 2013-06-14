package revel

import (
	"compress/gzip"
	"compress/zlib"
	"net/http"
	"strings"
)

type WriteFlusher interface {
	Write([]byte) (int, error)
	Flush() error
}

type CompressedResponseWriter struct {
	http.ResponseWriter
	w       WriteFlusher
	written bool
}

var typePreference = [...]string{
	"gzip",
	"deflate",
}

func CompressionFilter(c *Controller, fc []Filter) {
	if Config.BoolDefault("results.compressed", false) {
		acceptedEncodings := c.Request.Request.Header.Get("Accept-Encoding")

		for _, encoding := range typePreference {
			if strings.Contains(acceptedEncodings, encoding) {
				var writer WriteFlusher

				switch encoding {
				case "gzip":
					writer = gzip.NewWriter(c.Response.Out)
				case "deflate":
					writer = zlib.NewWriter(c.Response.Out)
				}

				c.Response.Out.Header().Set("Content-Encoding", encoding)
				c.Response.Out = &CompressedResponseWriter{c.Response.Out, writer, false}
				break
			}
		}
	}
	fc[0](c, fc[1:])
}

func (c *CompressedResponseWriter) Write(b []byte) (int, error) {
	if !c.written {
		c.Header().Del("Content-Length")
		c.written = true
	}
	num, err := c.w.Write(b)
	c.w.Flush()
	return num, err
}
