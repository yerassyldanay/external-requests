package postgresconn

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/yerassyldanay/requestmaker/pkg/configx"
)

func NewDB(cfg configx.Configuration) (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		cfg.PostgresHostname, cfg.PostgresPort, cfg.PostgresUsername, cfg.PostgresDBName, cfg.PostgresPassword, cfg.PostgresSSLMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
