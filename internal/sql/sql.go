package sql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
)

func Handle(data string) error {
	conn, err := connect()
	defer conn.Close()
	if err != nil {
		return err
	}

	tableName := os.Getenv("SQL_SERVER_TABLE")

	data = strings.TrimRight(data, "\n")
	arr := strings.Split(data, "\n")
	for _, line := range arr {
		fields := strings.Split(line, " ")
		if len(fields) != 3 {
			log.Println(line)
			continue
		}
		if err := insertUrl(conn, tableName, fields[0], fields[1], fields[2], true); err != nil {
			return err
		}
	}

	return nil
}

func connect() (*sql.DB, error) {
	server := os.Getenv("SQL_SERVER_NAME")
	user := os.Getenv("SQL_SERVER_USER")
	password := os.Getenv("SQL_SERVER_PASSWORD")
	database := os.Getenv("SQL_SERVER_DATABASE")
	port := "1433"
	if v := os.Getenv("SQL_SERVER_PORT"); v != "" {
		port = v
	}

	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", server, user, password, port, database)
	return sql.Open("mssql", connString)
}

func insertUrl(db *sql.DB, tableName, apiVersion, url, action string, isManagementPlane bool) error {
	isManagementPlaneNum := 0
	if isManagementPlane {
		isManagementPlaneNum = 1
	}
	fullResourceName := "Unknown"
	if index := strings.LastIndex(url, "/PROVIDERS/"); index != -1 {
		fullResourceName = url[index+len("/PROVIDERS/"):]
		fullResourceName = strings.TrimRight(fullResourceName, "/")
	}
	tsql := fmt.Sprintf("insert into %s (APIVersion, OperationName, HttpMethod, IsManagementPlane, FullResourceName) values ('%s', '%s', '%s', %d, '%s');", tableName, apiVersion, url, action, isManagementPlaneNum, fullResourceName)
	if _, err := db.Exec(tsql); err != nil {
		fmt.Println("Error inserting new row: " + err.Error())
		return err
	}
	return nil
}
