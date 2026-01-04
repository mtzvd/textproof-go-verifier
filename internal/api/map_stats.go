package api

import (
	"strconv"

	"blockchain-verifier/internal/viewmodels"
	"blockchain-verifier/web/templates/components"
)

// mapStatsCards преобразует данные статистики в параметры для отображения в карточках
func mapStatsCards(stats viewmodels.StatsResponse) []components.StatsCardParams {
	last := "—"
	if !stats.LastAdded.IsZero() {
		last = stats.LastAdded.Format("02.01.2006 15:04")
	}

	return []components.StatsCardParams{
		{
			Icon:     "fas fa-file-alt",
			Title:    "Всего текстов",
			Value:    strconv.Itoa(stats.TotalBlocks),
			Subtitle: "Зафиксированных документов",
		},
		{
			Icon:     "fas fa-users",
			Title:    "Уникальных авторов",
			Value:    strconv.Itoa(stats.UniqueAuthors),
			Subtitle: "Авторов зарегистрировано",
		},
		{
			Icon:     "fas fa-calendar-check",
			Title:    "Последнее депонирование",
			Value:    last,
			Subtitle: "Дата и время последней фиксации",
		},
	}
}
