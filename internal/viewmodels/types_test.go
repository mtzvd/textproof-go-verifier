package viewmodels

import (
	"testing"
	"time"
)

// types_test.go

func TestDepositRequest(t *testing.T) {
	req := DepositRequest{
		AuthorName: "John Doe",
		Title:      "Test Article",
		Text:       "Test content",
		PublicKey:  "test-key",
	}

	if req.AuthorName != "John Doe" {
		t.Error("AuthorName not set")
	}
	if req.Title != "Test Article" {
		t.Error("Title not set")
	}
	if req.Text != "Test content" {
		t.Error("Text not set")
	}
	if req.PublicKey != "test-key" {
		t.Error("PublicKey not set")
	}
}

func TestDepositResponse(t *testing.T) {
	now := time.Now()
	resp := DepositResponse{
		ID:        "000-000-001",
		Hash:      "abc123",
		Timestamp: now,
		QRCodeURL: "http://example.com/qr/000-000-001",
		BadgeURL:  "http://example.com/badge/000-000-001",
		Duplicate: false,
	}

	if resp.ID != "000-000-001" {
		t.Error("ID incorrect")
	}
	if resp.Hash != "abc123" {
		t.Error("Hash incorrect")
	}
	if resp.Timestamp != now {
		t.Error("Timestamp incorrect")
	}
	if resp.Duplicate {
		t.Error("Duplicate should be false")
	}
}

func TestVerifyByIDRequest(t *testing.T) {
	req := VerifyByIDRequest{ID: "000-000-001"}
	if req.ID != "000-000-001" {
		t.Error("ID not set correctly")
	}
}

func TestVerifyByTextRequest(t *testing.T) {
	req := VerifyByTextRequest{Text: "test text"}
	if req.Text != "test text" {
		t.Error("Text not set correctly")
	}
}

func TestVerificationResponse(t *testing.T) {
	now := time.Now()
	resp := VerificationResponse{
		Found:     true,
		BlockID:   "000-000-001",
		Author:    "John Doe",
		Title:     "Article",
		Timestamp: now,
		Hash:      "abc123",
		Matches:   true,
	}

	if !resp.Found {
		t.Error("Found should be true")
	}
	if resp.BlockID != "000-000-001" {
		t.Error("BlockID incorrect")
	}
	if resp.Author != "John Doe" {
		t.Error("Author incorrect")
	}
	if !resp.Matches {
		t.Error("Matches should be true")
	}
}

func TestStatsResponse(t *testing.T) {
	now := time.Now()
	resp := StatsResponse{
		TotalBlocks:   100,
		UniqueAuthors: 25,
		LastAdded:     now,
		ChainValid:    true,
	}

	if resp.TotalBlocks != 100 {
		t.Error("TotalBlocks incorrect")
	}
	if resp.UniqueAuthors != 25 {
		t.Error("UniqueAuthors incorrect")
	}
	if !resp.ChainValid {
		t.Error("ChainValid should be true")
	}
}

func TestBlockchainInfoResponse(t *testing.T) {
	resp := BlockchainInfoResponse{
		Length:     50,
		Difficulty: 4,
		Valid:      true,
		LastBlock:  "000-000-050",
	}

	if resp.Length != 50 {
		t.Error("Length incorrect")
	}
	if resp.Difficulty != 4 {
		t.Error("Difficulty incorrect")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse{
		Error:   "error message",
		Details: "error details",
	}

	if resp.Error != "error message" {
		t.Error("Error message incorrect")
	}
	if resp.Details != "error details" {
		t.Error("Details incorrect")
	}
}

func TestFlashData(t *testing.T) {
	flash := FlashData{
		Show:        true,
		Type:        "success",
		Message:     "Success message",
		IsDuplicate: false,
	}

	if !flash.Show {
		t.Error("Show should be true")
	}
	if flash.Type != "success" {
		t.Error("Type incorrect")
	}
	if flash.IsDuplicate {
		t.Error("IsDuplicate should be false")
	}
}

func TestDepositResultVM(t *testing.T) {
	now := time.Now()
	vm := DepositResultVM{
		ID:        "000-000-001",
		Title:     "Test Title",
		Author:    "Test Author",
		Hash:      "abc123",
		Timestamp: now,
		QRCodeURL: "/qr/000-000-001",
		BadgeURL:  "/badge/000-000-001",
		VerifyURL: "/verify/000-000-001",
	}

	if vm.ID != "000-000-001" {
		t.Error("ID incorrect")
	}
	if vm.Title != "Test Title" {
		t.Error("Title incorrect")
	}
	if vm.Author != "Test Author" {
		t.Error("Author incorrect")
	}
	if vm.Hash != "abc123" {
		t.Error("Hash incorrect")
	}
	if vm.VerifyURL != "/verify/000-000-001" {
		t.Error("VerifyURL incorrect")
	}
}

func TestZeroValues(t *testing.T) {
	var req DepositRequest
	if req.AuthorName != "" {
		t.Error("Zero value AuthorName should be empty")
	}

	var resp DepositResponse
	if resp.Duplicate {
		t.Error("Zero value Duplicate should be false")
	}

	var flash FlashData
	if flash.Show {
		t.Error("Zero value Show should be false")
	}
}

func TestStructCopy(t *testing.T) {
	original := DepositRequest{
		AuthorName: "Original",
		Title:      "Title",
		Text:       "Text",
	}

	copy := original
	copy.AuthorName = "Modified"

	if original.AuthorName != "Original" {
		t.Error("Original struct was modified")
	}
	if copy.AuthorName != "Modified" {
		t.Error("Copy struct was not modified")
	}
}

func BenchmarkDepositRequest_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DepositRequest{
			AuthorName: "Author",
			Title:      "Title",
			Text:       "Text",
			PublicKey:  "Key",
		}
	}
}
