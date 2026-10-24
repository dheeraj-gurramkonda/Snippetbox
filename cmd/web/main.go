package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"time"

	"snippetbox.dheeraj.net/internal/models"

	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"

	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

// application holds the dependencies shared by HTTP handlers, so handlers
// don't need to manage database, logging, template, or session setup themselves.
type application struct {
    logger        *slog.Logger
    snippets       models.SnippetModelInterface 
    users          models.UserModelInterface    
    templateCache  map[string]*template.Template
    formDecoder    *form.Decoder
    sessionManager *scs.SessionManager
}

func main() {
	// Command-line flags let deployments choose the listening address and database.
	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Open and verify the database connection before serving requests.
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	// Parse templates once at startup; requests can reuse the prepared cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	// Persist sessions in MySQL and expire them after twelve hours.
	sessionManager := scs.New()
	sessionManager.Store = mysqlstore.New(db)
	sessionManager.Lifetime = 12 * time.Hour

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DB: db},
		users:          &models.UserModel{DB: db},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// Require modern TLS and bound how long connections may remain idle or
	// spend reading and writing a request.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}

	srv := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", *addr)

	// ListenAndServeTLS blocks until the server stops or encounters an error.
	err = srv.ListenAndServeTLS("./assets/tls/cert.pem", "./assets/tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

// openDB creates a database handle and verifies that the configured database
// can be reached before returning it to the application.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		// sql.Open is lazy, so close the handle if the first real connection
		// check fails.
		db.Close()
		return nil, err
	}

	return db, nil
}
