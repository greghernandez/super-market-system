package db

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{DB: db}
}

// Create inserts a new record
func (r *BaseRepository) Create(model interface{}) error {
	result := r.DB.Create(model)
	if result.Error != nil {
		return fmt.Errorf("failed to create record: %w", result.Error)
	}
	return nil
}

// GetByID retrieves a record by ID
func (r *BaseRepository) GetByID(model interface{}, id string) error {
	result := r.DB.Where("id = ?", id).First(model)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return errors.New("record not found")
	}
	if result.Error != nil {
		return fmt.Errorf("failed to get record: %w", result.Error)
	}
	return nil
}

// Update updates a record
func (r *BaseRepository) Update(model interface{}, updates interface{}) error {
	result := r.DB.Model(model).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("failed to update record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}
	return nil
}

// Delete soft deletes a record by ID
func (r *BaseRepository) Delete(model interface{}, id string) error {
	result := r.DB.Where("id = ?", id).Delete(model)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}
	return nil
}

// List retrieves records with optional filters
func (r *BaseRepository) List(models interface{}, conditions ...interface{}) error {
	query := r.DB
	
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	
	result := query.Find(models)
	if result.Error != nil {
		return fmt.Errorf("failed to list records: %w", result.Error)
	}
	return nil
}

// Count returns the number of records matching the conditions
func (r *BaseRepository) Count(model interface{}, conditions ...interface{}) (int64, error) {
	var count int64
	query := r.DB.Model(model)
	
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	
	result := query.Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count records: %w", result.Error)
	}
	return count, nil
}

// Exists checks if a record exists with the given conditions
func (r *BaseRepository) Exists(model interface{}, conditions ...interface{}) (bool, error) {
	var count int64
	query := r.DB.Model(model)
	
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	
	result := query.Limit(1).Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("failed to check existence: %w", result.Error)
	}
	return count > 0, nil
}

// Transaction executes a function within a database transaction
func (r *BaseRepository) Transaction(fn func(*gorm.DB) error) error {
	return r.DB.Transaction(fn)
}