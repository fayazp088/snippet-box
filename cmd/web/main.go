package main

import (
	"database/sql"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"

	"github.com/fayazp088/snippet-box/internal/models"
	"github.com/go-playground/form"
	_ "github.com/go-sql-driver/mysql"
)

type Application struct {
	logger       *slog.Logger
	snippets     *models.SnippetModel
	templteCache map[string]*template.Template
	formDecoder  *form.Decoder
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	addr := flag.String("addr", ":8080", "HTTP network address")

	dsn := flag.String("dsn", "admin:admin@/snippets?parseTime=true", "MYSQL Data Source Name")
	db, err := OpenDB(*dsn)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	tmplCache, err := templateCache()

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	app := &Application{
		logger:       logger,
		snippets:     &models.SnippetModel{DB: db},
		templteCache: tmplCache,
		formDecoder:  formDecoder,
	}

	flag.Parse()

	logger.Info("starting server on :8080", "addr", *addr)
	err = http.ListenAndServe(*addr, app.routes())

	if err != nil {
		logger.Error(err.Error())
	}

	os.Exit(1)
}

func OpenDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
