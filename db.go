package main

import (
	"database/sql"
	"fmt"
	"hash"
	"time"

	log "github.com/sirupsen/logrus"

	// Needed implicitly to enable Postgres driver
	_ "github.com/lib/pq"
)

// Database defines methods to be implemented by SQLdb
type Database interface {
	GetChecksum(file File) error
	Close()
}

// SQLdb struct that acts as a receiver for the DB update methods
type SQLdb struct {
	DB       *sql.DB
	ConnInfo string
}

// DBConfig stores information about the database backend
type DBConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	Database   string
	CACert     string
	SslMode    string
	ClientCert string
	ClientKey  string
}

// dbRetryTimes is the number of times to retry the same function if it fails
var dbRetryTimes = 8

// dbReconnectTimeout is how long to try to re-establish a connection to the database
var dbReconnectTimeout = 5 * time.Minute

// dbReconnectSleep is how long to wait between attempts to connect to the database
var dbReconnectSleep = 5 * time.Second

// sqlOpen is an internal variable to ease testing
var sqlOpen = sql.Open

// logFatalf is an internal variable to ease testing
var logFatalf = log.Fatalf

// hashType returns the identification string for the hash type
func hashType(h hash.Hash) string {
	// TODO: Support/check type
	return "SHA256"
}

// NewDB creates a new DB connection
func NewDB(config DBConfig) (*SQLdb, error) {
	connInfo := buildConnInfo(config)

	log.Debugf("Connecting to DB %s:%d on database: %s with user: %s", config.Host, config.Port, config.Database, config.User)
	db, err := sqlOpen("postgres", connInfo)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return &SQLdb{DB: db, ConnInfo: connInfo}, nil
}

// buildConnInfo builds a connection string for the database
func buildConnInfo(config DBConfig) string {
	connInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.Database, config.SslMode)

	if config.SslMode == "disable" {
		return connInfo
	}

	if config.CACert != "" {
		connInfo += fmt.Sprintf(" sslrootcert=%s", config.CACert)
	}

	if config.ClientCert != "" {
		connInfo += fmt.Sprintf(" sslcert=%s", config.ClientCert)
	}

	if config.ClientKey != "" {
		connInfo += fmt.Sprintf(" sslkey=%s", config.ClientKey)
	}

	return connInfo
}

// checkAndReconnectIfNeeded validates the current connection with a ping
// and tries to reconnect if necessary
func (dbs *SQLdb) checkAndReconnectIfNeeded() {
	start := time.Now()

	for dbs.DB.Ping() != nil {
		log.Errorln("Database unreachable, reconnecting")
		dbs.DB.Close()

		if time.Since(start) > dbReconnectTimeout {
			logFatalf("Could not reconnect to failed database in reasonable time, giving up")
		}
		time.Sleep(dbReconnectSleep)
		log.Debugln("Reconnecting to DB")
		dbs.DB, _ = sqlOpen("postgres", dbs.ConnInfo)
	}

}

// GetChecksum retrieves the file header
func (dbs *SQLdb) GetChecksum(file File) error {
	var (
		err   error = nil
		count int   = 0
	)

	for count == 0 || (err != nil && count < dbRetryTimes) {
		err = dbs.getChecksum(file)
		count++
	}
	return err
}

// getChecksum is the actual function performing work for GetHeader
func (dbs *SQLdb) getChecksum(file File) error {
	dbs.checkAndReconnectIfNeeded()

	db := dbs.DB
	const query = "SELECT archive_file_checksum, archive_file_checksum_type FROM local_ega.main WHERE submission_file_path =  $1"

	var checksum, checksumType string
	if err := db.QueryRow(query, file.FileName).Scan(&checksum, &checksumType); err != nil {
		return err
	}

	return nil
}

// Close terminates the connection to the database
func (dbs *SQLdb) Close() {
	db := dbs.DB
	db.Close()
}
