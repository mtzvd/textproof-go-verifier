package config

import (
	"testing"
)

func TestGetAppConstants(t *testing.T) {
	constants := GetAppConstants()

	if constants.AppName == "" {
		t.Error("AppName should not be empty")
	}

	if constants.AppVersion == "" {
		t.Error("AppVersion should not be empty")
	}

	if constants.GitHubURL == "" {
		t.Error("GitHubURL should not be empty")
	}

	if constants.GitHubIssuesURL == "" {
		t.Error("GitHubIssuesURL should not be empty")
	}

	if constants.ContactEmail == "" {
		t.Error("ContactEmail should not be empty")
	}

	if constants.LegalEmail == "" {
		t.Error("LegalEmail should not be empty")
	}

	if constants.PrivacyEmail == "" {
		t.Error("PrivacyEmail should not be empty")
	}
}

func TestGitHubRepoURL(t *testing.T) {
	url := GitHubRepoURL()
	if url == "" {
		t.Error("GitHubRepoURL should not be empty")
	}
	if url != GetAppConstants().GitHubURL {
		t.Errorf("GitHubRepoURL() = %s, want %s", url, GetAppConstants().GitHubURL)
	}
}

func TestGitHubIssues(t *testing.T) {
	url := GitHubIssues()
	if url == "" {
		t.Error("GitHubIssues should not be empty")
	}
	if url != GetAppConstants().GitHubIssuesURL {
		t.Errorf("GitHubIssues() = %s, want %s", url, GetAppConstants().GitHubIssuesURL)
	}
}

func TestContactEmailAddress(t *testing.T) {
	email := ContactEmailAddress()
	if email == "" {
		t.Error("ContactEmailAddress should not be empty")
	}
	if email != GetAppConstants().ContactEmail {
		t.Errorf("ContactEmailAddress() = %s, want %s", email, GetAppConstants().ContactEmail)
	}
}

func TestLegalEmailAddress(t *testing.T) {
	email := LegalEmailAddress()
	if email == "" {
		t.Error("LegalEmailAddress should not be empty")
	}
	if email != GetAppConstants().LegalEmail {
		t.Errorf("LegalEmailAddress() = %s, want %s", email, GetAppConstants().LegalEmail)
	}
}

func TestEmailPrivacy(t *testing.T) {
	email := EmailPrivacy()
	if email == "" {
		t.Error("EmailPrivacy should not be empty")
	}
	if email != GetAppConstants().PrivacyEmail {
		t.Errorf("EmailPrivacy() = %s, want %s", email, GetAppConstants().PrivacyEmail)
	}
}

func TestConstants_Consistency(t *testing.T) {
	constants := GetAppConstants()

	if constants.AppName != "TextProof" {
		t.Logf("AppName = %s (expected TextProof)", constants.AppName)
	}

	if constants.AppVersion == "" || constants.AppVersion == "0.0.0" {
		t.Log("AppVersion might need to be updated")
	}
}

func BenchmarkGetAppConstants(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetAppConstants()
	}
}

func BenchmarkConvenienceMethods(b *testing.B) {
	b.Run("GitHubRepoURL", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GitHubRepoURL()
		}
	})

	b.Run("GitHubIssues", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GitHubIssues()
		}
	})

	b.Run("ContactEmailAddress", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ContactEmailAddress()
		}
	})
}
