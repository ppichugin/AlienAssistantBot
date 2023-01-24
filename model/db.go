package model

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ppichugin/AlienAssistantBot/config"
)

func NewDB(tries uint) error {
	if tries == 0 {
		return errors.New("tries can not be zero")
	}
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.GlobConf.HostDB,
		config.GlobConf.PortDB,
		config.GlobConf.UserDB,
		config.GlobConf.PasswordDB,
		config.GlobConf.NameDB)
	var db *sql.DB
	var err error
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
	return fmt.Errorf(fmt.Sprintf("Failed to connect to DB after %d tries (%s)", tries, err))
}
