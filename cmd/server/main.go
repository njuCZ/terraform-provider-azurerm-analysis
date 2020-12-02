package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"log"
	"net/http"
	"os"
	"os/exec"

	"gopkg.in/robfig/cron.v3"
)

var sem = semaphore.NewWeighted(1)

func main() {
	log.Println("Create new cron")
	c := cron.New()
	// every Sunday execute once
	c.AddFunc("0 0 0 ? * 1", func(){
		log.Println("cron job triggered")
		if _, err := OnlyOneRefresh(); err != nil {
			log.Printf("%+v\n", err)
			return
		}
		log.Println("cron job success")
	})
	c.Start()

	http.HandleFunc("/trigger", func(w http.ResponseWriter, req *http.Request){
		output, err := OnlyOneRefresh()
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
		fmt.Fprintf(w, "%s", output)
	})
	http.ListenAndServe(":8080", nil)
}

func OnlyOneRefresh() (string, error) {
	if !sem.TryAcquire(1) {
		return "", fmt.Errorf("refreshing")
	}
	defer sem.Release(1)
	return refresh()
}

func refresh() (string, error){
	//step 1: git pull azurerm repo
	cmd := exec.CommandContext(context.Background(), "git", "pull")
	cmd.Dir = os.Getenv("PROVIDER_REPO_PATH")

	if err := cmd.Run(); err != nil {
		return "", err
	}

	//step 2: execute extract
	extractCmdPath := os.Getenv("EXTRACT_CMD_PATH")
	cmd = exec.CommandContext(context.Background(), extractCmdPath, "./...")
	cmd.Dir = os.Getenv("PROVIDER_REPO_PATH")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
