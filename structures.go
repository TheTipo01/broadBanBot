package main

type Config struct {
	Token    string `fig:"token" validate:"required"`
	LogLevel string `fig:"loglevel"`
	Driver   string `fig:"driver" validate:"required"`
	DSN      string `fig:"dsn" validate:"required"`
}
