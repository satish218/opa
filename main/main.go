package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/open-policy-agent/opa/rego"
)

type Input struct {
	Method string   `json:"method"`
	Path   []string `json:"path"`
}

func main() {
	ctx := context.Background()

	// Load the Rego policy from the file
	policyFile := "policy.rego"
	policy, err := os.ReadFile(policyFile)
	if err != nil {
		log.Fatalf("Failed to read policy file: %v", err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// input data
		input := Input{
			Method: r.Method,
			Path:   []string{strings.TrimPrefix(r.URL.Path, "/")},
		}

		// new OPA client with the loaded policy
		client := rego.New(
			rego.Query("data.example.allow"),
			rego.Module(policyFile, string(policy)),
		)

		// query for evaluation
		preparedQuery, err := client.PrepareForEval(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to prepare policy for evaluation: %v", err), http.StatusInternalServerError)
			return
		}

		// evaluating the policy with the input data
		rs, err := preparedQuery.Eval(ctx, rego.EvalInput(input))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to evaluate policy: %v", err), http.StatusInternalServerError)
			return
		}

		// checking the policy decision to respond accordingly
		if len(rs) > 0 && rs[0].Expressions[0].Value.(bool) {
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
