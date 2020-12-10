package db

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/njucz/terraform-provider-azurerm-analysis/internal/common"
	"strings"
)

type EndpointDBManger struct {
	server   string
	database string
	port     string
	user     string
	password string
	table    string
	conn     *sql.DB
}

func NewEndpointDBManger(server, database, port, user, password, table string) (*EndpointDBManger, func() error, error) {
	manager := &EndpointDBManger{
		server:   server,
		database: database,
		port:     port,
		user:     user,
		password: password,
		table:    table,
	}
	conn, err := manager.connect()
	if err != nil {
		return nil, nil, err
	}
	manager.conn = conn
	return manager, conn.Close, nil
}

func (e *EndpointDBManger) connect() (*sql.DB, error) {
	connString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%s;database=%s;", e.server, e.user, e.password, e.port, e.database)
	return sql.Open("mssql", connString)
}

func (e *EndpointDBManger) Query(month string) ([]common.Endpoint, error) {
	stmt := fmt.Sprintf("select APIVersion, OperationName, HttpMethod, IsManagementPlane from %s where Month = ?;", e.table)
	rows, err := e.conn.Query(stmt, month)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	results := make([]common.Endpoint, 0)
	for rows.Next() {
		endpoint := common.Endpoint{}
		if err := rows.Scan(&endpoint.ApiVersion, &endpoint.Url, &endpoint.HttpMethod, &endpoint.IsManagementPlane); err != nil {
			return nil, err
		}
		results = append(results, endpoint)
	}
	return results, nil
}

func (e *EndpointDBManger) BatchInsertTransaction(month string, endpoints []common.Endpoint, size int) error {
	tx, err := e.conn.Begin()
	if err != nil {
		return err
	}

	if err := e.batchInsertLimitSize(month, endpoints, size); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (e *EndpointDBManger) batchInsertLimitSize(month string, endpoints []common.Endpoint, size int) error {
	left := 0
	for right := size; right < len(endpoints); right += size {
		if err := e.batchInsert(month, endpoints[left:right]); err != nil {
			return err
		}
		left = right
	}
	return e.batchInsert(month, endpoints[left:])
}

func (e *EndpointDBManger) batchInsert(month string, endpoints []common.Endpoint) error {
	valueStrings := make([]string, 0, len(endpoints))
	valueArgs := make([]interface{}, 0, len(endpoints)*5)
	for _, endpoint := range endpoints {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?, ?)")

		valueArgs = append(valueArgs, month)
		valueArgs = append(valueArgs, endpoint.ApiVersion)
		valueArgs = append(valueArgs, endpoint.Url)
		valueArgs = append(valueArgs, endpoint.HttpMethod)
		valueArgs = append(valueArgs, endpoint.IsManagementPlane)
		valueArgs = append(valueArgs, endpoint.GetFullResourceName())
	}

	stmt := fmt.Sprintf("insert into %s (Month, APIVersion, OperationName, HttpMethod, IsManagementPlane, FullResourceName) values %s;", e.table, strings.Join(valueStrings, ","))
	if _, err := e.conn.Exec(stmt, valueArgs...); err != nil {
		return fmt.Errorf("batch insert: %+v", err)
	}
	return nil
}
