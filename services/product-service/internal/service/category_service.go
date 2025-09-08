package service

import (
	"errors"
	"fmt"
	"product-service/internal/models"
	"product-service/internal/repository"
)

type CategoryService struct {
	repo repository.ProductRepository
}

func NewCategoryService(repo repository.ProductRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

func (s *CategoryService) GetCategory(id string) (*models.Category, error) {
	if id == "" {
		return nil, errors.New("category ID is required")
	}

	category, err := s.repo.GetCategory(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) ListCategories(parentID *string) ([]models.Category, error) {
	categories, err := s.repo.ListCategories(parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}

func (s *CategoryService) CreateCategory(request *models.CreateCategoryRequest) (*models.Category, error) {
	if request == nil {
		return nil, errors.New("create category request is required")
	}

	// Validate parent category if provided
	if request.ParentID != nil && *request.ParentID != "" {
		_, err := s.repo.GetCategory(*request.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
	}

	category := &models.Category{
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		ParentID:    request.ParentID,
	}

	err := s.repo.CreateCategory(category)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

func (s *CategoryService) UpdateCategory(id string, request *models.UpdateCategoryRequest) (*models.Category, error) {
	if id == "" {
		return nil, errors.New("category ID is required")
	}

	if request == nil {
		return nil, errors.New("update category request is required")
	}

	// Verify category exists
	_, err := s.repo.GetCategory(id)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Validate parent category if being updated
	if request.ParentID != nil && *request.ParentID != "" {
		if *request.ParentID == id {
			return nil, errors.New("category cannot be its own parent")
		}
		_, err := s.repo.GetCategory(*request.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parent category not found: %w", err)
		}
	}

	updatedCategory, err := s.repo.UpdateCategory(id, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return updatedCategory, nil
}

func (s *CategoryService) DeleteCategory(id string) error {
	if id == "" {
		return errors.New("category ID is required")
	}

	// Verify category exists
	_, err := s.repo.GetCategory(id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Soft delete by setting is_active = false
	isActive := false
	_, err = s.repo.UpdateCategory(id, &models.UpdateCategoryRequest{
		IsActive: &isActive,
	})

	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}