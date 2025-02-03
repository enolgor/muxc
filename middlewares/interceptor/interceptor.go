package interceptor

import (
	"bytes"
	"io"
	"net/http"
)

func Copy(status int, headers http.Header, body io.Reader, rw http.ResponseWriter) {
	CopyHeaders(headers, rw)
	CopyStatus(status, rw)
	CopyBody(body, rw)
}

func CopyStatus(status int, rw http.ResponseWriter) {
	rw.WriteHeader(status)
}

func CopyHeaders(headers http.Header, rw http.ResponseWriter) {
	for k, v := range headers {
		for _, h := range v {
			rw.Header().Add(k, h)
		}
	}
}

func CopyBody(body io.Reader, rw http.ResponseWriter) {
	io.Copy(rw, body)
}

func New(interceptor func(status int, headers http.Header, body io.Reader, rw http.ResponseWriter, req *http.Request)) func(next http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, req *http.Request) {
			brw := newBufferedResponseWriter()
			next.ServeHTTP(brw, req)
			brw.seal()
			brw.body.Seek(0, 0)
			interceptor(brw.status, brw.header.Clone(), brw.body, w, req)
		}
	}
}

type bufferedResponseWriter struct {
	header      http.Header
	buffer      *bytes.Buffer
	status      int
	wroteHeader bool
	body        *bytes.Reader
}

func newBufferedResponseWriter() *bufferedResponseWriter {
	return &bufferedResponseWriter{
		header:      make(map[string][]string),
		buffer:      &bytes.Buffer{},
		status:      http.StatusOK,
		wroteHeader: false,
	}
}

func (brw *bufferedResponseWriter) Header() http.Header {
	return brw.header
}

func (brw *bufferedResponseWriter) Write(data []byte) (int, error) {
	if !brw.wroteHeader {
		brw.wroteHeader = true
	}
	return brw.buffer.Write(data)
}

func (brw *bufferedResponseWriter) WriteHeader(statusCode int) {
	if !brw.wroteHeader {
		brw.status = statusCode
	}
	brw.wroteHeader = true
}

func (brw *bufferedResponseWriter) seal() {
	brw.body = bytes.NewReader(brw.buffer.Bytes())
}
