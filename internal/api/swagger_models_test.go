package api

import (
	"testing"

	"blockchain-verifier/internal/viewmodels"
)

// swagger_models_test.go

func TestSwaggerModels_DepositRequest(t *testing.T) {
	req := viewmodels.DepositRequest{
		AuthorName: "John Doe",
		Title:      "Test Article",
		Text:       "Test content",
		PublicKey:  "test-key",
	}

	if req.AuthorName != "John Doe" {
		t.Error("AuthorName not set correctly")
	}
	if req.Title != "Test Article" {
		t.Error("Title not set correctly")
	}
	if req.Text != "Test content" {
		t.Error("Text not set correctly")
	}
	if req.PublicKey != "test-key" {
		t.Error("PublicKey not set correctly")
	}
}

func TestSwaggerModels_DepositResponse(t *testing.T) {
	resp := viewmodels.DepositResponse{
		ID:        "000-000-001",
		Hash:      "abc123",
		QRCodeURL: "http://localhost/qr/000-000-001",
		BadgeURL:  "http://localhost/badge/000-000-001",
		Duplicate: false,
	}

	if resp.ID != "000-000-001" {
		t.Error("ID not set correctly")
	}
	if resp.Hash != "abc123" {
		t.Error("Hash not set correctly")
	}
	if resp.Duplicate {
		t.Error("Duplicate should be false")
	}
}

func TestSwaggerModels_VerifyByIDRequest(t *testing.T) {
	req := viewmodels.VerifyByIDRequest{
		ID: "000-000-001",
	}

	if req.ID != "000-000-001" {
		t.Error("ID not set correctly")
	}
}

func TestSwaggerModels_VerifyByTextRequest(t *testing.T) {
	req := viewmodels.VerifyByTextRequest{
		Text: "Test text content",
	}

	if req.Text != "Test text content" {
		t.Error("Text not set correctly")
	}
}

func TestSwaggerModels_VerificationResponse(t *testing.T) {
	resp := viewmodels.VerificationResponse{
		Found:   true,
		BlockID: "000-000-001",
		Author:  "John Doe",
		Title:   "Test",
		Matches: true,
	}

	if !resp.Found {
		t.Error("Found should be true")
	}
	if !resp.Matches {
		t.Error("Matches should be true")
	}
	if resp.Author != "John Doe" {
		t.Error("Author not set correctly")
	}
}

func TestSwaggerModels_ErrorResponse(t *testing.T) {
	resp := viewmodels.ErrorResponse{
		Error: "Test error message",
	}

	if resp.Error != "Test error message" {
		t.Error("Error message not set correctly")
	}
}

func TestSwaggerModels_StatsResponse(t *testing.T) {
	resp := viewmodels.StatsResponse{
		TotalBlocks:   100,
		UniqueAuthors: 25,
		ChainValid:    true,
	}

	if resp.TotalBlocks != 100 {
		t.Error("TotalBlocks not set correctly")
	}
	if resp.UniqueAuthors != 25 {
		t.Error("UniqueAuthors not set correctly")
	}
	if !resp.ChainValid {
		t.Error("ChainValid should be true")
	}
}

func TestSwaggerModels_BlockchainInfoResponse(t *testing.T) {
	resp := viewmodels.BlockchainInfoResponse{
		Length:     50,
		Difficulty: 4,
		Valid:      true,
		LastBlock:  "000-000-050",
	}

	if resp.Length != 50 {
		t.Error("Length not set correctly")
	}
	if resp.Difficulty != 4 {
		t.Error("Difficulty not set correctly")
	}
	if !resp.Valid {
		t.Error("Valid should be true")
	}
}

func TestSwaggerModels_ZeroValues(t *testing.T) {
	// Тестируем что zero values работают корректно
	var req viewmodels.DepositRequest
	if req.AuthorName != "" {
		t.Error("Zero value should be empty string")
	}

	var resp viewmodels.DepositResponse
	if resp.Duplicate {
		t.Error("Zero value for Duplicate should be false")
	}
}
