package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/refresh", refresh)

	http.ListenAndServe(":8080", nil)
}

func refresh(w http.ResponseWriter, req *http.Request) {
	//todo 1: git pull azurerm repo
	cmd := exec.CommandContext(context.Background(), "git", "pull")
	cmd.Dir = os.Getenv("PROVIDER_REPO_PATH")

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	//todo 2: execute extract
	extractCmdPath := os.Getenv("EXTRACT_CMD_PATH")
	cmd = exec.CommandContext(context.Background(), extractCmdPath, "./...")
	cmd.Dir = os.Getenv("PROVIDER_REPO_PATH")

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
	}

	//todo 3: print output
	fmt.Fprintf(w, "%s", string(output))
}
