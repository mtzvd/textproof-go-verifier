package main

import (
	"blockchain-verifier/internal/api"
	"blockchain-verifier/internal/blockchain"
	"blockchain-verifier/internal/config"
	"context"
	"fmt"
	"log"
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
// @contact.name   Georgiy Agafonov
// @contact.email  info@web-n-roll.pl
// @contact.url    https://github.com/mtzvd/textproof-go-verifier
//
// @license.name  MIT
// @license.url   https://github.com/mtzvd/textproof-go-verifier/blob/main/LICENSE
//
// @host      localhost:8080
// @BasePath  /
//
// @schemes   http https
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
		log.Fatalf("Ошибка конфигурации: %v", err)
	}

	// Выводим информацию о конфигурации
	fmt.Println("=== TextProof Blockchain ===")
	fmt.Printf("Директория данных: %s\n", cfg.DataDir)
	fmt.Printf("Порт сервера: %d\n", cfg.Port)
	fmt.Printf("Сложность майнинга: %d нулей\n", cfg.Difficulty)
	fmt.Printf("Режим отладки: %v\n", cfg.EnableDebug)
	fmt.Println()

	// Создаем хранилище
	storage, err := blockchain.NewStorage(cfg.DataDir)
	if err != nil {
		log.Fatalf("Не удалось создать хранилище: %v", err)
	}

	// Создаем блокчейн
	bc, err := blockchain.NewBlockchain(storage, cfg.Difficulty)
	if err != nil {
		log.Fatalf("Не удалось создать блокчейн: %v", err)
	}

	// Выводим информацию о цепочке
	info := bc.GetChainInfo()
	fmt.Println("=== Информация о блокчейне ===")
	fmt.Printf("Длина цепочки: %d блоков\n", info["length"])
	fmt.Printf("Цепочка валидна: %v\n", info["valid"])

	if lastBlock, ok := info["last_block"]; ok {
		fmt.Printf("Последний блок: %s\n", lastBlock)
	}
	fmt.Println()

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
				fmt.Println("=== Запуск тестового сценария в фоне ===")

				if err := runTestScenario(bc); err != nil {
					log.Printf("Тестовый сценарий не удался: %v", err)
				} else {
					fmt.Println("✓ Тестовый сценарий успешно выполнен")
				}
				fmt.Println()
			}
		}()
	}

	// Канал для системных сигналов
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Запуск HTTP сервера в горутине
	go func() {
		fmt.Printf("=== Запуск HTTP сервера на http://localhost:%d ===\n", cfg.Port)
		fmt.Println("Доступные страницы:")
		fmt.Printf("  • Главная: http://localhost:%d/\n", cfg.Port)
		fmt.Printf("  • Депонирование: http://localhost:%d/deposit\n", cfg.Port)
		fmt.Printf("  • Проверка: http://localhost:%d/verify\n", cfg.Port)
		fmt.Printf("  • Swagger UI: http://localhost:%d/swagger/index.html\n", cfg.Port)
		fmt.Println("\nНажмите Ctrl+C для остановки")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Не удалось запустить сервер: %v", err)
		}
	}()

	// Ожидание сигнала остановки
	<-stop
	fmt.Println("\n=== Получен сигнал остановки ===")

	// Graceful shutdown с таймаутом 10 секунд
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Останавливаем сервер...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	}

	fmt.Println("✓ Сервер остановлен")
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
		fmt.Printf("Добавление тестового блока %d/%d...\n", i+1, len(testDeposits))
		_, err := bc.AddBlock(data)
		if err != nil {
			return fmt.Errorf("не удалось добавить блок: %w", err)
		}
	}

	return nil
}
