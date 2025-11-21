package internal

import (
	"fmt"
	"io"
	"os"
)

type config struct {
	Prefix string
	Host   string
	Port   string
	LogTo  io.Writer
}

func NewConfig() *config {
	return &config{
		Prefix: "UNOGAME_SERVICE",
	}
}

func (c *config) Load() error {
	c.Host = os.Getenv(fmt.Sprintf("%s_HOST", c.Prefix))
	c.Port = os.Getenv(fmt.Sprintf("%s_PORT", c.Prefix))

	logfile := os.Getenv(fmt.Sprintf("%s_LOG_FILE", c.Prefix))

	if c.Port == "" {
		c.Port = "8000"
	}

	if logfile != "" {
		file, err := os.OpenFile(logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return err
		}
		c.LogTo = file
	} else {
		c.LogTo = os.Stdout
	}

	return nil
}

func (c *config) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
