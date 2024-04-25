package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", healthCheck)

	fmt.Println("Statistics service is on...")
	http.ListenAndServe(":8080", nil)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Health is ok.")
}
