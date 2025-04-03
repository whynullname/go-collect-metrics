package postgres

import (
	"database/sql"

	"github.com/whynullname/go-collect-metrics/internal/logger"
)

type Postgres struct {
	DB     *sql.DB
	Adress string
}

func NewPostgresRepo(adress string) *Postgres {
	//ps := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", `localhost`, `videos`, `XXXXXXXX`, `videos`)
	db, err := sql.Open("pgx", adress)
	if err != nil {
		logger.Log.Error(err)
		return nil
	}
	//defer db.Close()

	return &Postgres{
		DB:     db,
		Adress: adress,
	}
}

func (p *Postgres) CloseRepo() {
	p.DB.Close()
}
