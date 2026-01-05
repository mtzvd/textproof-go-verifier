package config

// AppConstants содержит константы приложения
type AppConstants struct {
	AppName         string
	AppVersion      string
	GitHubURL       string
	GitHubIssuesURL string
	ContactEmail    string
	LegalEmail      string
	PrivacyEmail    string
}

// GetAppConstants возвращает константы приложения
func GetAppConstants() AppConstants {
	return AppConstants{
		AppName:         "TextProof",
		AppVersion:      "1.0.0",
		GitHubURL:       "https://github.com/mtzvd/textproof-go-verifier",
		GitHubIssuesURL: "https://github.com/mtzvd/textproof-go-verifier/issues",
		ContactEmail:    "privacy@textproof.example.com",
		LegalEmail:      "legal@textproof.example.com",
		PrivacyEmail:    "privacy@textproof.example.com",
	}
}

// Convenience методы для быстрого доступа

// GitHubRepoURL возвращает URL репозитория
func GitHubRepoURL() string {
	return GetAppConstants().GitHubURL
}

// GitHubIssues возвращает URL для создания issue
func GitHubIssues() string {
	return GetAppConstants().GitHubIssuesURL
}

// ContactEmailAddress возвращает email для связи
func ContactEmailAddress() string {
	return GetAppConstants().ContactEmail
}

// LegalEmailAddress возвращает email для юридических вопросов
func LegalEmailAddress() string {
	return GetAppConstants().LegalEmail
}

// EmailPrivacy возвращает email для вопросов по конфиденциальности
func EmailPrivacy() string {
	return GetAppConstants().PrivacyEmail
}
