package compressmiddleware

import (
	"compress/gzip"
	"io"
	"net/http"

	"github.com/whynullname/go-collect-metrics/internal/logger"
)

type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GZIP(next http.Handler) http.Handler {
	acceptedContentTypes := []string{
		"text/html",
		"application/json",
	}

	acceptedContentEncoding := "gzip"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptType := r.Header.Values("Accept")
		hasAsseptedContentType := false
		for _, contentType := range acceptType {
			for _, asseptedContentType := range acceptedContentTypes {
				if contentType == asseptedContentType {
					hasAsseptedContentType = true
					break
				}
			}
		}

		if !hasAsseptedContentType {
			next.ServeHTTP(w, r)
			return
		}

		ow := w

		acceptEncoding := r.Header.Values("Accept-Encoding")
		for _, encodingType := range acceptEncoding {
			if encodingType == acceptedContentEncoding {
				cw := newCompressWriter(w)
				ow = cw
				defer cw.Close()
				break
			}
		}

		contentEncoding := r.Header.Values("Content-Encoding")
		for _, encodingType := range contentEncoding {
			if encodingType == acceptedContentEncoding {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logger.Log.Infof("error %s", err.Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				r.Body = cr
				defer cr.Close()
				break
			}
		}

		next.ServeHTTP(ow, r)
	})
}
