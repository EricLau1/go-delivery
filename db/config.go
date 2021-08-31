package db

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config interface {
	URI() string
}

type config struct {
	user string
	pass string
	host string
	port int
	name string
}

func NewConfig() Config {

	cfg := &config{
		user: os.Getenv("DB_USER"),
		pass: os.Getenv("DB_PASS"),
		host: os.Getenv("DB_HOST"),
		name: os.Getenv("DB_NAME"),
	}

	var err error
	cfg.port, err = strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		log.Println(err)
	}

	return cfg
}

func (cfg *config) URI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?w=majority", cfg.user, cfg.pass, cfg.host, cfg.port, cfg.name)
}
