package database

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ems/config"

	"github.com/jmoiron/sqlx"
	"github.com/tursodatabase/go-libsql"
)

var DB *sqlx.DB

var replica *libsql.Connector

const driver = "libsql"

func Connect(cfg *config.Config) {
	target := strings.TrimSpace(cfg.TursoURL)
	if target == "" {
		log.Fatal("Failed to connect to database: TURSO_DATABASE_URL is empty")
	}

	if strings.TrimSpace(cfg.TursoReplicaPath) != "" {
		connectReplica(cfg, target)
	} else {
		connectDirect(cfg, target)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	DB.SetMaxOpenConns(1)

	fmt.Printf("Database connected: %s\n", describe(target))
}

func connectDirect(cfg *config.Config, target string) {
	dsn := normalizeDSN(target)
	if isLocal(dsn) {
		ensureLocalDir(dsn)
	} else {
		dsn = withAuthToken(dsn, cfg.TursoAuthToken)
	}

	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB = db

	if isLocal(dsn) {
		configureLocal(db)
	}
}

func connectReplica(cfg *config.Config, target string) {
	dsn := normalizeDSN(target)
	if isLocal(dsn) {
		log.Fatal("Database configuration error: TURSO_REPLICA_PATH requires a remote TURSO_DATABASE_URL")
	}

	primary, token := remoteAndToken(dsn, cfg.TursoAuthToken)
	replicaPath := strings.TrimPrefix(cfg.TursoReplicaPath, "file:")
	ensureLocalDir("file:" + replicaPath)

	opts := make([]libsql.Option, 0, 2)
	if token != "" {
		opts = append(opts, libsql.WithAuthToken(token))
	}
	if cfg.TursoSyncInterval > 0 {
		opts = append(opts, libsql.WithSyncInterval(time.Duration(cfg.TursoSyncInterval)*time.Second))
	}

	connector, err := libsql.NewEmbeddedReplicaConnector(replicaPath, primary, opts...)
	if err != nil {
		log.Fatal("Failed to open embedded replica:", err)
	}
	replica = connector
	DB = sqlx.NewDb(sql.OpenDB(connector), driver)

	configureLocal(DB)
}

func Close() {
	if DB != nil {
		_ = DB.Close()
		DB = nil
	}
	if replica != nil {
		_ = replica.Close()
		replica = nil
	}
}

// normalizeDSN accepts a bare file path for local databases and adds the
// `file:` scheme libSQL expects. Remote URLs are returned unchanged.
func normalizeDSN(target string) string {
	lower := strings.ToLower(target)
	switch {
	case strings.HasPrefix(lower, "file:"),
		strings.HasPrefix(lower, ":memory:"),
		strings.HasPrefix(lower, "libsql://"),
		strings.HasPrefix(lower, "http://"),
		strings.HasPrefix(lower, "https://"):
		return target
	default:
		return "file:" + target
	}
}

func isLocal(dsn string) bool {
	lower := strings.ToLower(dsn)
	return strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, ":memory:")
}

// withAuthToken adds Turso's authToken query parameter without overriding one
// already supplied in TURSO_DATABASE_URL.
func withAuthToken(dsn, token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return dsn
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	query := u.Query()
	if query.Get("authToken") != "" {
		return dsn
	}
	query.Set("authToken", token)
	u.RawQuery = query.Encode()
	return u.String()
}

func remoteAndToken(dsn, token string) (string, string) {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn, strings.TrimSpace(token)
	}

	if strings.TrimSpace(token) == "" {
		token = u.Query().Get("authToken")
	}
	query := u.Query()
	query.Del("authToken")
	u.RawQuery = query.Encode()
	return u.String(), strings.TrimSpace(token)
}

func ensureLocalDir(dsn string) {
	if strings.EqualFold(dsn, ":memory:") {
		return
	}

	path := strings.TrimPrefix(dsn, "file:")
	if i := strings.IndexByte(path, '?'); i >= 0 {
		path = path[:i]
	}
	if path == "" || path == ":memory:" {
		return
	}

	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatal("Failed to create database directory:", err)
		}
	}
}

func configureLocal(db *sqlx.DB) {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		"PRAGMA foreign_keys = ON",
	}
	for _, pragma := range pragmas {
		rows, err := db.Query(pragma)
		if err != nil {
			log.Printf("Database pragma %q failed: %v", pragma, err)
			continue
		}
		_ = rows.Close()
	}
}

// describe returns a log-safe target (never prints the auth token).
func describe(target string) string {
	u, err := url.Parse(target)
	if err != nil || u.Scheme == "" {
		return "local file " + strings.TrimPrefix(target, "file:")
	}
	if u.Scheme == "file" {
		return "local file " + strings.TrimPrefix(target, "file:")
	}
	query := u.Query()
	query.Del("authToken")
	u.RawQuery = query.Encode()
	return u.String()
}
