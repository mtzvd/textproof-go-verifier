package config

import (
	"flag"
	"fmt"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	DataDir     string
	Port        int
	Difficulty  int
	EnableDebug bool
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		DataDir:     "data",
		Port:        8080,
		Difficulty:  4,
		EnableDebug: false,
	}
}

// LoadFromFlags загружает конфигурацию из флагов командной строки
func (c *Config) LoadFromFlags() {
	flag.StringVar(&c.DataDir, "data-dir", c.DataDir, "Директория для хранения данных")
	flag.IntVar(&c.Port, "port", c.Port, "Порт для HTTP сервера")
	flag.IntVar(&c.Difficulty, "difficulty", c.Difficulty, "Сложность майнинга (количество нулей)")
	flag.BoolVar(&c.EnableDebug, "debug", c.EnableDebug, "Включить режим отладки")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: %s [опции]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Опции:")
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "\nПримеры:")
		fmt.Fprintln(os.Stderr, "  server -data-dir ./my_data -port 9090")
		fmt.Fprintln(os.Stderr, "  server -difficulty 3 -debug")
	}

	flag.Parse()
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	if c.Difficulty < 1 || c.Difficulty > 6 {
		return fmt.Errorf("сложность должна быть от 1 до 6")
	}
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("порт должен быть от 1 до 65535")
	}
	return nil
}
