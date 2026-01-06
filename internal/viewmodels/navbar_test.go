package viewmodels

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNavBarItem(t *testing.T) {
	item := NavBarItem{
		Label:  "Home",
		Href:   "/",
		Active: true,
		Icon:   "fas fa-home",
		Align:  "start",
	}

	if item.Label != "Home" {
		t.Error("Label not set correctly")
	}
	if item.Href != "/" {
		t.Error("Href not set correctly")
	}
	if !item.Active {
		t.Error("Active should be true")
	}
	if item.Icon != "fas fa-home" {
		t.Error("Icon not set correctly")
	}
	if item.Align != "start" {
		t.Error("Align not set correctly")
	}
}

func TestNavBarItem_WithoutIcon(t *testing.T) {
	item := NavBarItem{
		Label:  "Page",
		Href:   "/page",
		Active: false,
	}

	if item.Icon != "" {
		t.Error("Icon should be empty")
	}
}

func TestNavBarItem_InactiveByDefault(t *testing.T) {
	var item NavBarItem
	item.Label = "Test"
	item.Href = "/test"

	if item.Active {
		t.Error("Active should be false by default")
	}
}

func TestNavBar(t *testing.T) {
	items := []NavBarItem{
		{Label: "Home", Href: "/", Active: true},
		{Label: "About", Href: "/about", Active: false},
		{Label: "Contact", Href: "/contact", Active: false},
	}

	navbar := NavBar{
		Brand: "TestApp",
		Icon:  "fas fa-test",
		Items: items,
	}

	if len(navbar.Items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(navbar.Items))
	}

	if navbar.Brand != "TestApp" {
		t.Error("Brand not set correctly")
	}

	if navbar.Icon != "fas fa-test" {
		t.Error("Icon not set correctly")
	}

	if !navbar.Items[0].Active {
		t.Error("First item should be active")
	}
}

func TestNavBar_EmptyItems(t *testing.T) {
	navbar := NavBar{
		Brand: "TestApp",
		Items: []NavBarItem{},
	}

	if len(navbar.Items) != 0 {
		t.Error("Items should be empty")
	}
}

func TestBuildHomeNavBar_HomePage(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	navbar := BuildHomeNavBar(req)

	if len(navbar.Items) == 0 {
		t.Error("Navbar should have items")
	}

	if navbar.Brand != "TextProof" {
		t.Errorf("Brand = %s, want TextProof", navbar.Brand)
	}

	if navbar.Icon != "fas fa-fingerprint" {
		t.Errorf("Icon = %s, want fas fa-fingerprint", navbar.Icon)
	}

	hasActive := false
	for _, item := range navbar.Items {
		if item.Active && item.Href == "/" {
			hasActive = true
			break
		}
	}

	if !hasActive {
		t.Error("Home page item should be active")
	}
}

func TestBuildHomeNavBar_DifferentPages(t *testing.T) {
	pages := []string{"/", "/deposit", "/verify", "/about", "/docs"}

	for _, page := range pages {
		t.Run("page="+page, func(t *testing.T) {
			req := httptest.NewRequest("GET", page, nil)
			navbar := BuildHomeNavBar(req)

			if len(navbar.Items) == 0 {
				t.Error("Navbar should have items")
			}

			foundActive := false
			for _, item := range navbar.Items {
				if item.Active && item.Href == page {
					foundActive = true
					break
				}
			}

			if !foundActive && page != "/verify" {
				t.Errorf("Item for %s should be active", page)
			}
		})
	}
}

func TestBuildHomeNavBar_VerifySubpath(t *testing.T) {
	req := httptest.NewRequest("GET", "/verify/abc123", nil)
	navbar := BuildHomeNavBar(req)

	foundActive := false
	for _, item := range navbar.Items {
		if item.Label == "Проверить" && item.Active {
			foundActive = true
			break
		}
	}

	if !foundActive {
		t.Error("Verify item should be active for /verify subpaths")
	}
}

func TestBuildHomeNavBar_InvalidPage(t *testing.T) {
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	navbar := BuildHomeNavBar(req)

	if len(navbar.Items) == 0 {
		t.Error("Navbar should have items even for invalid page")
	}
}

func TestBuildHomeNavBar_RequiredPages(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	navbar := BuildHomeNavBar(req)

	requiredPages := []string{"/", "/deposit", "/verify", "/about", "/docs"}

	for _, required := range requiredPages {
		found := false
		for _, item := range navbar.Items {
			if item.Href == required {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Navbar should contain link to %s", required)
		}
	}
}

func TestBuildHomeNavBar_ActiveLogic(t *testing.T) {
	req := httptest.NewRequest("GET", "/deposit", nil)
	navbar := BuildHomeNavBar(req)

	for _, item := range navbar.Items {
		if item.Href == "/deposit" {
			if !item.Active {
				t.Error("/deposit item should be active")
			}
		} else if item.Href == "/" || item.Href == "/about" || item.Href == "/docs" {
			if item.Active {
				t.Errorf("%s item should not be active", item.Href)
			}
		}
	}
}

func TestBuildHomeNavBar_Consistency(t *testing.T) {
	req1 := httptest.NewRequest("GET", "/", nil)
	req2 := httptest.NewRequest("GET", "/", nil)

	navbar1 := BuildHomeNavBar(req1)
	navbar2 := BuildHomeNavBar(req2)

	if len(navbar1.Items) != len(navbar2.Items) {
		t.Error("BuildHomeNavBar should return consistent results")
	}

	for i := range navbar1.Items {
		if navbar1.Items[i].Label != navbar2.Items[i].Label {
			t.Error("Items order should be consistent")
		}
		if navbar1.Items[i].Href != navbar2.Items[i].Href {
			t.Error("Items should be identical")
		}
	}
}

func TestBuildHomeNavBar_IconsPresent(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	navbar := BuildHomeNavBar(req)

	for _, item := range navbar.Items {
		if item.Icon == "" {
			t.Logf("Item %s has no icon", item.Label)
		}
	}
}

func TestBuildHomeNavBar_AlignmentValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	navbar := BuildHomeNavBar(req)

	for _, item := range navbar.Items {
		if item.Align != "start" && item.Align != "end" && item.Align != "" {
			t.Errorf("Invalid align value: %s", item.Align)
		}
	}
}

func BenchmarkBuildHomeNavBar(b *testing.B) {
	req := httptest.NewRequest("GET", "/", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHomeNavBar(req)
	}
}

func BenchmarkBuildHomeNavBar_DifferentPages(b *testing.B) {
	pages := []string{"/", "/deposit", "/verify", "/about", "/docs"}
	requests := make([]*http.Request, len(pages))

	for i, page := range pages {
		requests[i] = httptest.NewRequest("GET", page, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildHomeNavBar(requests[i%len(requests)])
	}
}
