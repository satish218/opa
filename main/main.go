package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Input struct {
	Method string   `json:"method"`
	Path   []string `json:"path"`
}

type OPARequest struct {
	Input Input `json:"input"`
}

type OPAResponse struct {
	Result bool `json:"result"`
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// input data
		input := Input{
			Method: r.Method,
			Path:   []string{strings.TrimPrefix(r.URL.Path, "/")},
		}

		// create a request to OPA
		opaReq := OPARequest{Input: input}
		opaReqJson, err := json.Marshal(opaReq)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to prepare OPA request: %v", err), http.StatusInternalServerError)
			return
		}

		resp, err := http.Post("http://localhost:8181/v1/data/example/allow", "application/json", strings.NewReader(string(opaReqJson)))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to send request to OPA: %v", err), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var opaResp OPAResponse
		if err := json.NewDecoder(resp.Body).Decode(&opaResp); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse response from OPA: %v", err), http.StatusInternalServerError)
			return
		}

		// checking the policy decision to respond accordingly
		if opaResp.Result {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Access granted\n"))
		} else {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Access denied\n"))
		}
	})

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
