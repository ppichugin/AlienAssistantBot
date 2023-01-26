package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // imports for DB API

	"github.com/ppichugin/AlienAssistantBot/config"
)

var ErrTriesZero = errors.New("tries can not be zero value")

func NewDB(tries uint) error {
	if tries == 0 {
		return ErrTriesZero
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.GlobConf.HostDB,
		config.GlobConf.PortDB,
		config.GlobConf.UserDB,
		config.GlobConf.PasswordDB,
		config.GlobConf.NameDB)

	var (
		db  *sql.DB
		err error
	)

	for i := tries; i > 0; i-- {
		time.Sleep(time.Second * 1)

		db, err = sql.Open("postgres", psqlInfo)
		if err != nil {
			log.Println(err)

			_ = db.Close()

			continue
		}

		if err = db.Ping(); err != nil {
			log.Println(err)

			_ = db.Close()

			continue
		}

		// Connected
		config.GlobConf.Database = db

		return nil
	}

	return fmt.Errorf("%w: failed to connect to DB after %d tries", err, tries)
}
