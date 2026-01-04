package api

import (
	"net/http"

	"blockchain-verifier/internal/viewmodels"
)

// FlashMessage представляет временное сообщение для пользователя
type FlashMessage struct {
	Type    string // "success", "warning", "info", "danger"
	Message string
	Data    map[string]string // дополнительные данные
}

// setFlash устанавливает flash cookie
func setFlash(w http.ResponseWriter, flashType, message string, data map[string]string) {
	// Создаём cookie который живёт только до следующего запроса
	cookie := &http.Cookie{
		Name:     "flash_type",
		Value:    flashType,
		Path:     "/",
		MaxAge:   60, // 60 секунд (достаточно для редиректа)
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	cookie = &http.Cookie{
		Name:     "flash_message",
		Value:    message,
		Path:     "/",
		MaxAge:   60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// Если есть дополнительные данные (например, duplicate=true)
	if data != nil {
		for key, value := range data {
			cookie = &http.Cookie{
				Name:     "flash_" + key,
				Value:    value,
				Path:     "/",
				MaxAge:   60,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(w, cookie)
		}
	}
}

// getFlash читает и удаляет flash cookie
func getFlash(r *http.Request, w http.ResponseWriter) *FlashMessage {
	typeCookie, err := r.Cookie("flash_type")
	if err != nil {
		return nil // нет flash сообщения
	}

	messageCookie, err := r.Cookie("flash_message")
	if err != nil {
		return nil
	}

	flash := &FlashMessage{
		Type:    typeCookie.Value,
		Message: messageCookie.Value,
		Data:    make(map[string]string),
	}

	// Читаем дополнительные данные
	for _, cookie := range r.Cookies() {
		if len(cookie.Name) > 6 && cookie.Name[:6] == "flash_" && cookie.Name != "flash_type" && cookie.Name != "flash_message" {
			key := cookie.Name[6:] // убираем префикс "flash_"
			flash.Data[key] = cookie.Value
		}
	}

	// Удаляем все flash cookies после чтения
	deleteFlashCookies(w, r)

	return flash
}

// deleteFlashCookies удаляет все flash cookies
func deleteFlashCookies(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if len(cookie.Name) >= 6 && cookie.Name[:6] == "flash_" {
			deleteCookie := &http.Cookie{
				Name:     cookie.Name,
				Value:    "",
				Path:     "/",
				MaxAge:   -1, // удалить немедленно
				HttpOnly: true,
			}
			http.SetCookie(w, deleteCookie)
		}
	}
}

// FlashData структура для передачи в шаблон
type FlashData struct {
	Show        bool
	Type        string
	Message     string
	IsDuplicate bool
}

// getFlashData извлекает flash данные для шаблона
func getFlashData(r *http.Request, w http.ResponseWriter) viewmodels.FlashData {
	flash := getFlash(r, w)
	if flash == nil {
		return viewmodels.FlashData{Show: false}
	}

	isDuplicate := flash.Data["duplicate"] == "true"

	return viewmodels.FlashData{
		Show:        true,
		Type:        flash.Type,
		Message:     flash.Message,
		IsDuplicate: isDuplicate,
	}
}
