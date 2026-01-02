package blockchain

import (
	"fmt"
	"strconv"
	"strings"
)

// incrementID увеличивает ID блока
// Формат: "000-000-000" или "A-000-000-000"
func incrementID(id string) (string, error) {
	// Разделяем ID на компоненты
	parts := strings.Split(id, "-")

	// Если 4 части, значит есть буква
	if len(parts) == 4 {
		return incrementIDWithLetter(parts[0], parts[1], parts[2], parts[3])
	}

	// Если 3 части, значит без буквы
	if len(parts) == 3 {
		return incrementIDNoLetter(parts[0], parts[1], parts[2])
	}

	return "", fmt.Errorf("invalid ID format: %s", id)
}

// incrementIDNoLetter обрабатывает ID без буквы
func incrementIDNoLetter(part1, part2, part3 string) (string, error) {
	// Конвертируем в числа
	num1, err := strconv.Atoi(part1)
	if err != nil {
		return "", err
	}

	num2, err := strconv.Atoi(part2)
	if err != nil {
		return "", err
	}

	num3, err := strconv.Atoi(part3)
	if err != nil {
		return "", err
	}

	// Инкрементируем третью часть
	num3++

	// Обрабатываем переносы
	if num3 == 1000 {
		num3 = 0
		num2++
	}

	if num2 == 1000 {
		num2 = 0
		num1++
	}

	// Проверяем, не достигли ли миллиарда
	if num1 == 1000 {
		// Достигли миллиарда, добавляем букву A
		return "A-000-000-000", nil
	}

	// Форматируем обратно
	return fmt.Sprintf("%03d-%03d-%03d", num1, num2, num3), nil
}

// incrementIDWithLetter обрабатывает ID с буквой
func incrementIDWithLetter(letter, part1, part2, part3 string) (string, error) {
	// Конвертируем в числа
	num1, err := strconv.Atoi(part1)
	if err != nil {
		return "", err
	}

	num2, err := strconv.Atoi(part2)
	if err != nil {
		return "", err
	}

	num3, err := strconv.Atoi(part3)
	if err != nil {
		return "", err
	}

	// Инкрементируем третью часть
	num3++

	// Обрабатываем переносы
	if num3 == 1000 {
		num3 = 0
		num2++
	}

	if num2 == 1000 {
		num2 = 0
		num1++
	}

	if num1 == 1000 {
		num1 = 0
		// Увеличиваем букву
		if letter == "Z" {
			// Если достигли Z, возвращаемся к A (или можно выбрасывать ошибку)
			letter = "A"
		} else {
			letter = string(letter[0] + 1)
		}
	}

	// Форматируем обратно
	return fmt.Sprintf("%s-%03d-%03d-%03d", letter, num1, num2, num3), nil
}

// isValidID проверяет корректность формата ID
func isValidID(id string) bool {
	// Реализуем при необходимости
	return true
}
