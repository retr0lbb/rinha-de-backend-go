package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handleHello)

	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleHello(w http.ResponseWriter, _ *http.Request) {
	wc, err := w.Write([]byte("Hello, World!\n"))

	if err != nil {
		slog.Error("Error Writing the response", "err", err)
		return
	}

	fmt.Printf("%d bytes written\n", wc)
}
