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
)

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
		fmt.Println("\nНажмите Ctrl+C для остановки")
		fmt.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	<-stop
	fmt.Println("\n=== Остановка сервера ===")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	} else {
		fmt.Println("✓ Сервер корректно остановлен")
	}
}

// runTestScenario выполняет тестовый сценарий
func runTestScenario(bc *blockchain.Blockchain) error {
	fmt.Println("=== Запуск тестового сценария ===")

	info := bc.GetChainInfo()
	if info["length"].(int) > 1 {
		fmt.Println("В цепочке уже есть блоки. Пропускаем тестовое добавление.")
		return nil
	}

	fmt.Println("1. Добавляем тестовый блок...")

	data1 := blockchain.DepositData{
		AuthorName:  "Иван Иванов",
		Title:       "Мой первый пост",
		TextStart:   "Это начало моего",
		TextEnd:     "конец моего текста",
		ContentHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		PublicKey:   "ssh-rsa AAAAB3NzaC1yc2E...",
	}

	start := time.Now()
	block1, err := bc.AddBlock(data1)
	if err != nil {
		return fmt.Errorf("не удалось добавить первый блок: %v", err)
	}

	fmt.Printf("   ✓ Блок добавлен за %v\n", time.Since(start))
	fmt.Printf("   ID: %s\n", block1.ID)
	fmt.Printf("   Хеш: %s\n", block1.Hash)
	fmt.Printf("   Nonce: %d\n", block1.Nonce)

	fmt.Println("\n2. Добавляем второй тестовый блок...")

	data2 := blockchain.DepositData{
		AuthorName:  "Петр Петров",
		Title:       "Статья о блокчейне",
		TextStart:   "Блокчейн это технология",
		TextEnd:     "будущее за децентрализацией",
		ContentHash: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		PublicKey:   "",
	}

	start = time.Now()
	block2, err := bc.AddBlock(data2)
	if err != nil {
		return fmt.Errorf("не удалось добавить второй блок: %v", err)
	}

	fmt.Printf("   ✓ Блок добавлен за %v\n", time.Since(start))
	fmt.Printf("   ID: %s\n", block2.ID)
	fmt.Printf("   Хеш: %s\n", block2.Hash)
	fmt.Printf("   Nonce: %d\n", block2.Nonce)

	fmt.Println("\n✓ Тестовый сценарий завершен!")
	return nil
}
