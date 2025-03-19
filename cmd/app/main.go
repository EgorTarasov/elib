package main

import (
	"elib/internal/client"
	"elib/internal/view"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func sendFile(w http.ResponseWriter, f *os.File) error {
	fileInfo, err := f.Stat()
	if err != nil {
		return errors.New("Error getting file info: " + err.Error())
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	_, err = io.Copy(w, f)
	if err != nil {
		return errors.New("Error sending file:" + err.Error())
	}
	return nil
}

func main() {
	port := flag.String("p", "8080", "http server port")
	flag.Parse()
	if p, err := strconv.Atoi(*port); err != nil || p < 0 || p > 65000 {
		if p < 0 || p > 65000 {
			panic("invalid http port")
		} else {
			panic(err)
		}
	}

	if len(os.Args) < 0 {
		fmt.Println("Usage: app <login> <password>")
		return
	}
	loginRaw := os.Args[1]
	login, err := strconv.Atoi(loginRaw)
	if err != nil {
		panic(err)
	}
	pass := os.Args[2]

	elib := client.NewElibClient()

	if err = elib.Login(login, pass); err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, view.Pages, "index.html")
	})

	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		rawBookID := r.URL.Query().Get("bookId")
		bookID, err := strconv.Atoi(rawBookID)
		if err != nil {
			_, _ = w.Write([]byte("invalid book id"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// checking if book already downloaded
		filename := fmt.Sprintf("book_%d.pdf", bookID)
		f, err := os.Open(filename)
		if err == nil {
			defer f.Close()
			err = sendFile(w, f)
			if err != nil {
				slog.Error("error sending cached file:", "err", err.Error())
			}

			return
		}

		err = elib.DownloadDocument(bookID)
		if err != nil {
			_, _ = w.Write([]byte(err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		f, err = os.Open(filename)
		if err != nil {
			_, _ = w.Write([]byte("unable to open result file: " + err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		err = sendFile(w, f)
		if err != nil {
			slog.Error("failed to send file", "err", err.Error())
		}
	})
	slog.Info("starting http server on:", "port", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		panic(err)
	}
}
