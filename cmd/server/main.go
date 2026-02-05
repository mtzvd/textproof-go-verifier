package main

import (
	"blockchain-verifier/internal/api"
	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/config"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "blockchain-verifier/docs" // Swagger документация (будет сгенерирована)
)

// @title           TextProof API
// @version         1.0
// @description     API для системы подтверждения авторства текстов с использованием blockchain
// @description     TextProof - это система для регистрации и верификации авторства текстовых документов через блокчейн с Proof-of-Work.
//
// @contact.name   TextProof
// @contact.email  info@textproof.ru
// @contact.url    https://textproof.ru
//
// @license.name  MIT
// @license.url   https://github.com/mtzvd/textproof-go-verifier/blob/main/LICENSE
//
// @host      textproof.ru
// @BasePath  /
//
// @schemes   https
//
// @tag.name         Deposit
// @tag.description  Операции депонирования (регистрации) текстов в блокчейне
//
// @tag.name         Verify
// @tag.description  Операции проверки и верификации текстов
//
// @tag.name         Stats
// @tag.description  Статистика и информация о блокчейне
//
// @tag.name         Utils
// @tag.description  Вспомогательные утилиты (QR-коды, badges)
//
// #Endpoints
// @description     Доступные конечные точки API:
// @description     - POST /api/v1/deposit - Регистрация текста
// @description     - POST /api/v1/verify/id - Проверка по ID
// @description     - POST /api/v1/verify/text - Проверка по тексту
// @description     - GET /api/v1/stats - Статистика
//
// @externalDocs.description  GitHub Repository
// @externalDocs.url          https://github.com/mtzvd/textproof-go-verifier
func main() {
	// Загружаем конфигурацию
	cfg := config.DefaultConfig()
	cfg.LoadFromFlags()

	if err := cfg.Validate(); err != nil {
		slog.Error("Ошибка конфигурации", "error", err)
		os.Exit(1)
	}

	slog.Info("Конфигурация",
		"data_dir", cfg.DataDir,
		"port", cfg.Port,
		"difficulty", cfg.Difficulty,
		"debug", cfg.EnableDebug,
	)

	// Создаем хранилище
	storage, err := blockchain.NewStorage(cfg.DataDir)
	if err != nil {
		slog.Error("Не удалось создать хранилище", "error", err)
		os.Exit(1)
	}

	// Создаем блокчейн
	bc, err := blockchain.NewBlockchain(storage, cfg.Difficulty)
	if err != nil {
		slog.Error("Не удалось создать блокчейн", "error", err)
		os.Exit(1)
	}

	// Выводим информацию о цепочке
	info := bc.GetChainInfo()
	logAttrs := []any{
		"length", info["length"],
		"valid", info["valid"],
	}
	if lastBlock, ok := info["last_block"]; ok {
		logAttrs = append(logAttrs, "last_block", lastBlock)
	}
	slog.Info("Блокчейн загружен", logAttrs...)

	// Создаем API
	apiHandler := api.NewAPI(bc)

	// Настраиваем HTTP сервер
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      apiHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Запускаем тестовый сценарий в фоне, если включен debug
	if cfg.EnableDebug {
		go func() {
			time.Sleep(2 * time.Second)

			info := bc.GetChainInfo()
			if info["length"].(int) <= 1 {
				slog.Info("Запуск тестового сценария")

				if err := runTestScenario(bc); err != nil {
					slog.Error("Тестовый сценарий не удался", "error", err)
				} else {
					slog.Info("Тестовый сценарий успешно выполнен")
				}
			}
		}()
	}

	// Канал для системных сигналов
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск HTTP сервера в горутине
	go func() {
		slog.Info("HTTP сервер запущен", "addr", fmt.Sprintf("http://localhost:%d", cfg.Port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Не удалось запустить сервер", "error", err)
			os.Exit(1)
		}
	}()

	// Ожидание сигнала остановки
	<-stop
	slog.Info("Получен сигнал остановки")

	// Graceful shutdown с таймаутом 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Останавливаем сервер...")
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Ошибка при остановке сервера", "error", err)
	}

	slog.Info("Сервер остановлен")
}

// runTestScenario запускает тестовый сценарий для проверки работы блокчейна
func runTestScenario(bc *blockchain.Blockchain) error {
	// Тестовые данные
	testDeposits := []blockchain.DepositData{
		{
			AuthorName:  "Александр Пушкин",
			Title:       "Евгений Онегин",
			TextStart:   "Мой дядя самых",
			TextEnd:     "достойнейших правил когда",
			ContentHash: "abc123def456",
			PublicKey:   "",
		},
		{
			AuthorName:  "Лев Толстой",
			Title:       "Война и мир",
			TextStart:   "Все счастливые семьи",
			TextEnd:     "похожи друг на",
			ContentHash: "xyz789uvw012",
			PublicKey:   "",
		},
	}

	// Добавляем блоки
	for i, data := range testDeposits {
		slog.Info("Добавление тестового блока", "num", i+1, "total", len(testDeposits))
		_, err := bc.AddBlock(data)
		if err != nil {
			return fmt.Errorf("не удалось добавить блок: %w", err)
		}
	}

	return nil
}
