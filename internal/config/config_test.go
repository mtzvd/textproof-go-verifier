package config

import (
	"flag"
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}

	if cfg.Difficulty != 4 {
		t.Errorf("Difficulty = %d, want 4", cfg.Difficulty)
	}

	if cfg.DataDir != "data" {
		t.Errorf("DataDir = %s, want data", cfg.DataDir)
	}

	if cfg.EnableDebug != false {
		t.Errorf("EnableDebug = %v, want false", cfg.EnableDebug)
	}
}

func TestLoadFromFlags(t *testing.T) {
	// Сохраняем оригинальные args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	t.Run("default values", func(t *testing.T) {
		// Сбрасываем флаги
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if cfg.Port != 8080 {
			t.Errorf("Port = %d, want 8080", cfg.Port)
		}

		if cfg.Difficulty != 4 {
			t.Errorf("Difficulty = %d, want 4", cfg.Difficulty)
		}

		if cfg.DataDir != "data" {
			t.Errorf("DataDir = %s, want data", cfg.DataDir)
		}
	})

	t.Run("custom port", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-port", "9090"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if cfg.Port != 9090 {
			t.Errorf("Port = %d, want 9090", cfg.Port)
		}
	})

	t.Run("custom difficulty", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-difficulty", "5"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if cfg.Difficulty != 5 {
			t.Errorf("Difficulty = %d, want 5", cfg.Difficulty)
		}
	})

	t.Run("custom data dir", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-data-dir", "/tmp/test"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if cfg.DataDir != "/tmp/test" {
			t.Errorf("DataDir = %s, want /tmp/test", cfg.DataDir)
		}
	})

	t.Run("enable debug", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-debug"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if !cfg.EnableDebug {
			t.Error("EnableDebug = false, want true")
		}
	})

	t.Run("all custom", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		os.Args = []string{"cmd", "-port", "3000", "-difficulty", "3", "-data-dir", "/custom", "-debug"}

		cfg := DefaultConfig()
		cfg.LoadFromFlags()

		if cfg.Port != 3000 {
			t.Errorf("Port = %d, want 3000", cfg.Port)
		}
		if cfg.Difficulty != 3 {
			t.Errorf("Difficulty = %d, want 3", cfg.Difficulty)
		}
		if cfg.DataDir != "/custom" {
			t.Errorf("DataDir = %s, want /custom", cfg.DataDir)
		}
		if !cfg.EnableDebug {
			t.Error("EnableDebug = false, want true")
		}
	})
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name       string
		port       int
		difficulty int
		wantErr    bool
	}{
		{"valid config", 8080, 4, false},
		{"valid min difficulty", 3000, 1, false},
		{"valid max difficulty", 8080, 6, false},
		{"port too low", 0, 4, true},
		{"port negative", -1, 4, true},
		{"port too high", 70000, 4, true},
		{"difficulty too low", 8080, 0, true},
		{"difficulty negative", 8080, -1, true},
		{"difficulty too high", 8080, 7, true},
		{"both invalid", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Port:       tt.port,
				Difficulty: tt.difficulty,
				DataDir:    "data",
			}

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

