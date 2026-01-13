package multi_storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	
	"postgresus-backend/internal/config"
	"postgresus-backend/internal/util/encryption"

	"github.com/google/uuid"
)

type StorageOperator interface {
	SaveFile(
		ctx context.Context,
		encryptor encryption.FieldEncryptor,
		logger *slog.Logger,
		fileID uuid.UUID,
		file io.Reader,
	) error
	GetFile(encryptor encryption.FieldEncryptor, fileID uuid.UUID) (io.ReadCloser, error)
	DeleteFile(encryptor encryption.FieldEncryptor, fileID uuid.UUID) error
	TestConnection(encryptor encryption.FieldEncryptor) error
}

type MultiStorage struct {
	StorageID   uuid.UUID `json:"storageId"   gorm:"primaryKey;type:uuid;column:storage_id"`
	PrimaryID   uuid.UUID `json:"primaryId"   gorm:"column:primary_id;type:uuid;not null"`
	SecondaryID uuid.UUID `json:"secondaryId" gorm:"column:secondary_id;type:uuid;not null"`

	Primary   StorageOperator `json:"-" gorm:"-"`
	Secondary StorageOperator `json:"-" gorm:"-"`
}

func (m *MultiStorage) TableName() string {
	return "multi_storages"
}

func (m *MultiStorage) SaveFile(
	ctx context.Context,
	encryptor encryption.FieldEncryptor,
	logger *slog.Logger,
	fileID uuid.UUID,
	file io.Reader,
) error {
	if m.Primary == nil || m.Secondary == nil {
		return fmt.Errorf("multi-storage not properly initialized")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tempFilePath := filepath.Join(config.GetEnv().TempFolder, "multi_"+fileID.String())
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tempFilePath)
	}()

	_, err = io.Copy(tempFile, file)
	if err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	_ = tempFile.Close()

	primaryFile, err := os.Open(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to open temp file for primary: %w", err)
	}
	err = m.Primary.SaveFile(ctx, encryptor, logger, fileID, primaryFile)
	_ = primaryFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save to primary storage: %w", err)
	}

	secondaryFile, err := os.Open(tempFilePath)
	if err != nil {
		return fmt.Errorf("failed to open temp file for secondary: %w", err)
	}
	err = m.Secondary.SaveFile(ctx, encryptor, logger, fileID, secondaryFile)
	_ = secondaryFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save to secondary storage: %w", err)
	}

	return nil
}

func (m *MultiStorage) GetFile(
	encryptor encryption.FieldEncryptor,
	fileID uuid.UUID,
) (io.ReadCloser, error) {
	if m.Primary == nil {
		return nil, fmt.Errorf("multi-storage not properly initialized")
	}

	reader, err := m.Primary.GetFile(encryptor, fileID)
	if err == nil {
		return reader, nil
	}

	if m.Secondary != nil {
		reader, err = m.Secondary.GetFile(encryptor, fileID)
		if err == nil {
			return reader, nil
		}
	}

	return nil, fmt.Errorf("failed to get file from both storages: %w", err)
}

func (m *MultiStorage) DeleteFile(encryptor encryption.FieldEncryptor, fileID uuid.UUID) error {
	var primaryErr, secondaryErr error

	if m.Primary != nil {
		primaryErr = m.Primary.DeleteFile(encryptor, fileID)
	}

	if m.Secondary != nil {
		secondaryErr = m.Secondary.DeleteFile(encryptor, fileID)
	}

	if primaryErr != nil && secondaryErr != nil {
		return fmt.Errorf("failed to delete from both storages: primary: %v, secondary: %v", primaryErr, secondaryErr)
	}

	return nil
}

func (m *MultiStorage) Validate(encryptor encryption.FieldEncryptor) error {
	if m.PrimaryID == uuid.Nil {
		return fmt.Errorf("primary storage is required")
	}
	if m.SecondaryID == uuid.Nil {
		return fmt.Errorf("secondary storage is required")
	}
	if m.PrimaryID == m.SecondaryID {
		return fmt.Errorf("primary and secondary storage must be different")
	}
	return nil
}

func (m *MultiStorage) TestConnection(encryptor encryption.FieldEncryptor) error {
	if m.Primary == nil || m.Secondary == nil {
		return fmt.Errorf("multi-storage not properly initialized")
	}

	if err := m.Primary.TestConnection(encryptor); err != nil {
		return fmt.Errorf("primary storage connection failed: %w", err)
	}

	if err := m.Secondary.TestConnection(encryptor); err != nil {
		return fmt.Errorf("secondary storage connection failed: %w", err)
	}

	return nil
}

func (m *MultiStorage) HideSensitiveData() {}

func (m *MultiStorage) EncryptSensitiveData(encryptor encryption.FieldEncryptor) error {
	return nil
}

func (m *MultiStorage) Update(incoming *MultiStorage) {
	m.PrimaryID = incoming.PrimaryID
	m.SecondaryID = incoming.SecondaryID
}

func (m *MultiStorage) SetStorages(primary, secondary StorageOperator) {
	m.Primary = primary
	m.Secondary = secondary
}
