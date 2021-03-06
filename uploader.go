package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

var (
	// server
	r    = mux.NewRouter()
	ip   = "0.0.0.0"
	port = "8082"
)


type LogRecord struct {
	http.ResponseWriter
	status int
}

func (r *LogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *LogRecord) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
func (r *LogRecord) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("not a Hijacker")
}

func WrapLogger(f http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := &LogRecord{
			ResponseWriter: w,
			status:         200,
		}
		t := time.Now()
		f.ServeHTTP(record, r)
		logrus.Infof("%v %v => %v, %v", r.Method, r.RequestURI, record.status, time.Since(t))
	})
}

// exists returns whether the given file or directory exists or not
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func main() {
	r.Path("/upload").Methods("POST").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dest := r.Header.Get("Kube-Destination")
		createDir, _ := strconv.ParseBool(r.Header.Get("Kube-Create-Dir"))

		if dest == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("'Kube-Destination' header required."))
			return
		}

		dirName := filepath.Dir(dest)
		if !Exists(dirName) && createDir {
			// Try to create dir
			if err := os.MkdirAll(dirName, os.ModePerm); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("Cannot create dir: %v\n", err.Error())))
				return
			}
		}

		file, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Cannot create file: %v\n", err.Error())))
			return
		}
		written, err := io.Copy(file, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error while writing file: %v\n", err.Error())))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("File %v created. Written %v bytes.\n", dest, written)))
	})

	logrus.Info("Starting...")
	logrus.Infof("Listen %v:%v", ip, port)
	http.ListenAndServe(fmt.Sprintf("%v:%v", ip, port), WrapLogger(r))
}
