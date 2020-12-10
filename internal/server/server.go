package server

import (
	"context"
	"fmt"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/common"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/server/cmd"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/server/db"
	"golang.org/x/sync/semaphore"
	"os"
)

type Server struct {
	gitRepo   cmd.GitRepo
	extractor cmd.AzurermProviderUrlExtractor
	sem       *semaphore.Weighted
}

func NewServer(providerRepoPath, gitCmdPath, extractCmdPath string) *Server {
	return &Server{
		gitRepo: cmd.GitRepo{
			Dir:        providerRepoPath,
			GitCmdPath: gitCmdPath,
		},
		extractor: cmd.AzurermProviderUrlExtractor{
			ExtractCmdPath: extractCmdPath,
			ProviderDir:    providerRepoPath,
		},
		sem: semaphore.NewWeighted(1),
	}
}

func (ser *Server) OnlyOneRefresh() (string, error) {
	if !ser.sem.TryAcquire(1) {
		return "", fmt.Errorf("refreshing")
	}
	defer ser.sem.Release(1)
	return ser.refresh()
}

func (ser *Server) refresh() (string, error) {
	ctx := context.Background()

	// step 1: git pull azurerm repo
	if err := ser.gitRepo.Pull(ctx); err != nil {
		return "", err
	}

	// step 2: execute extract
	output, endpoints, err := ser.extractor.ExtractUrl(ctx)
	if err != nil {
		return "", err
	}

	// step 3: insert data
	ser.compareAndInsert(endpoints)
	return output, err
}

func (ser *Server) compareAndInsert(endpoints []common.Endpoint) error {
	server := os.Getenv("SQL_SERVER_NAME")
	user := os.Getenv("SQL_SERVER_USER")
	password := os.Getenv("SQL_SERVER_PASSWORD")
	database := os.Getenv("SQL_SERVER_DATABASE")
	port := "1433"
	if v := os.Getenv("SQL_SERVER_PORT"); v != "" {
		port = v
	}
	table := os.Getenv("SQL_SERVER_TABLE")

	// init
	manager, close, err := db.NewEndpointDBManger(server, database, port, user, password, table)
	if err != nil {
		return err
	}
	defer close()

	month := getFirstDateOfCurrentMonth()
	// query
	originalEndpoints, err := manager.Query(month)
	if err != nil {
		return err
	}
	// compare
	newEndpoints := diff(originalEndpoints, endpoints)
	if len(newEndpoints) == 0 {
		fmt.Println("no new endpoints found")
		return nil
	}
	fmt.Printf("new endpoints found:\n%s", endpointsToString(newEndpoints))

	// insert
	return manager.BatchInsertTransaction(month, newEndpoints, 100)
}
