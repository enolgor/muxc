package middlewares

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/enolgor/muxc/middlewares/header"
	"github.com/enolgor/muxc/middlewares/interceptor"
	"github.com/enolgor/muxc/middlewares/logger"
	"github.com/enolgor/muxc/middlewares/recoverer"
	"github.com/go-chi/chi/v5/middleware"
)

var RequestID func(http.Handler) http.Handler = middleware.RequestID
var SetHeader = middleware.SetHeader
var InterceptContentSniffer = interceptor.New(func(status int, headers http.Header, body io.Reader, rw http.ResponseWriter, req *http.Request) {
	fmt.Println("interceptor content sniffer")
	if headers.Get(header.ContentType) != "application/json" {
		interceptor.Copy(status, headers, body, rw)
		return
	}
	data := struct {
		Payload json.RawMessage `json:"payload"`
		Success bool            `json:"success"`
	}{}
	dec := json.NewDecoder(body)
	if err := dec.Decode(&data.Payload); err != nil {
		http.Error(rw, fmt.Sprintf(`{"success": false, "error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	if status >= 200 && status <= 299 {
		data.Success = true
	}
	enc := json.NewEncoder(rw)
	interceptor.CopyHeaders(headers, rw)
	interceptor.CopyStatus(status, rw)
	enc.Encode(data)
})

var InterceptErrorStatus = interceptor.New(func(status int, headers http.Header, body io.Reader, rw http.ResponseWriter, req *http.Request) {
	fmt.Println("interceptor error status")
	if status < 400 {
		interceptor.Copy(status, headers, body, rw)
		return
	}
	data := struct {
		Error     string `json:"error"`
		RequestId string `json:"request_id"`
	}{}
	errMsg, _ := io.ReadAll(body)
	data.Error = strings.TrimSpace(string(errMsg))
	data.RequestId = logger.GetRequestId(req)
	enc := json.NewEncoder(rw)
	interceptor.CopyHeaders(headers, rw)
	rw.Header().Set(header.ContentType, "application/json")
	interceptor.CopyStatus(status, rw)
	enc.Encode(data)
})

var Recover func(http.HandlerFunc) http.HandlerFunc = recoverer.New(func(panicked any, w http.ResponseWriter, req *http.Request) {
	debug.PrintStack()
	http.Error(w, "runtime panic", http.StatusInternalServerError)
})
