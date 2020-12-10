package main

import (
	"fmt"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/server"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
)

func main() {
	logFlags := log.LstdFlags | log.Lshortfile
	log.SetFlags(logFlags)

	providerRepoPath := os.Getenv("PROVIDER_REPO_PATH")
	extractCmdPath := os.Getenv("EXTRACT_CMD_PATH")
	gitCmdPath := "git"

	ser := server.NewServer(providerRepoPath, gitCmdPath, extractCmdPath)

	log.Println("starting cron job")
	c := cron.New()
	// every Sunday execute once
	c.AddFunc("0 0 0 ? * 1", func() {
		log.Println("cron job triggered")
		if _, err := ser.OnlyOneRefresh(); err != nil {
			log.Printf("%+v\n", err)
			return
		}
		log.Println("cron job success")
	})
	c.Start()

	log.Println("starting http server")
	http.HandleFunc("/trigger", func(w http.ResponseWriter, req *http.Request) {
		log.Println("http request triggered")
		defer log.Println("http request end")
		output, err := ser.OnlyOneRefresh()
		if err != nil {
			fmt.Fprintf(w, "%+v\n", err)
			return
		}
		fmt.Fprintf(w, "%s", output)
	})
	http.ListenAndServe(":8080", nil)
}
