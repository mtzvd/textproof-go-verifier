package blockchain

import "fmt"

// Ошибки блокчейна
type BlockchainError struct {
	Code    string
	Message string
	Err     error
}

func (e *BlockchainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *BlockchainError) Unwrap() error {
	return e.Err
}

// Конструкторы ошибок
func NewBlockchainError(code, message string, err error) *BlockchainError {
	return &BlockchainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Конкретные ошибки
var (
	ErrInvalidBlockHash = &BlockchainError{
		Code:    "INVALID_HASH",
		Message: "invalid block hash"}
	ErrInvalidDifficulty = &BlockchainError{
		Code:    "INVALID_DIFFICULTY",
		Message: "block doesn't meet difficulty requirement"}
	ErrPrevHashMismatch = &BlockchainError{
		Code:    "PREV_HASH_MISMATCH",
		Message: "previous hash doesn't match"}
	ErrInvalidIDFormat = &BlockchainError{
		Code:    "INVALID_ID_FORMAT",
		Message: "invalid ID format"}
	ErrChainValidationFailed = &BlockchainError{
		Code:    "CHAIN_VALIDATION_FAILED",
		Message: "blockchain validation failed"}
	ErrBlockNotFound = &BlockchainError{
		Code:    "BLOCK_NOT_FOUND",
		Message: "block not found"}
	ErrStorage = &BlockchainError{
		Code:    "STORAGE_ERROR",
		Message: "storage error"}
	ErrWALWriteFailed = &BlockchainError{
		Code:    "WAL_WRITE_FAILED",
		Message: "failed to write to WAL"}
	ErrWALRecoveryFailed = &BlockchainError{
		Code:    "WAL_RECOVERY_FAILED",
		Message: "failed to recover from WAL"}
	ErrBackupRestoreFailed = &BlockchainError{
		Code:    "BACKUP_RESTORE_FAILED",
		Message: "failed to restore from backup"}
	ErrDuplicateContentHash = &BlockchainError{
		Code:    "DUPLICATE_CONTENT_HASH",
		Message: "block with same content hash already exists",
	}
)

// DuplicateBlockError сигнализирует о попытке добавить дублирующий блок
type DuplicateBlockError struct {
	Block *Block
}

//
func (e *DuplicateBlockError) Error() string {
	return "block with this content hash already exists"
}
