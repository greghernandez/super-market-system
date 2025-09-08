package service

import (
	"errors"
	"fmt"
	"product-service/internal/models"
	"product-service/internal/repository"
)

type DepartmentService struct {
	repo repository.ProductRepository
}

func NewDepartmentService(repo repository.ProductRepository) *DepartmentService {
	return &DepartmentService{
		repo: repo,
	}
}

func (s *DepartmentService) GetDepartment(id string) (*models.Department, error) {
	if id == "" {
		return nil, errors.New("department ID is required")
	}

	department, err := s.repo.GetDepartment(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get department: %w", err)
	}

	return department, nil
}

func (s *DepartmentService) ListDepartments() ([]models.Department, error) {
	departments, err := s.repo.ListDepartments()
	if err != nil {
		return nil, fmt.Errorf("failed to list departments: %w", err)
	}

	return departments, nil
}

func (s *DepartmentService) CreateDepartment(request *models.CreateDepartmentRequest) (*models.Department, error) {
	if request == nil {
		return nil, errors.New("create department request is required")
	}

	department := &models.Department{
		Name:        request.Name,
		Description: request.Description,
		Icon:        request.Icon,
		Image:       request.Image,
		Slug:        request.Slug,
	}

	err := s.repo.CreateDepartment(department)
	if err != nil {
		return nil, fmt.Errorf("failed to create department: %w", err)
	}

	return department, nil
}

func (s *DepartmentService) UpdateDepartment(id string, request *models.UpdateDepartmentRequest) (*models.Department, error) {
	if id == "" {
		return nil, errors.New("department ID is required")
	}

	if request == nil {
		return nil, errors.New("update department request is required")
	}

	// Verify department exists
	_, err := s.repo.GetDepartment(id)
	if err != nil {
		return nil, fmt.Errorf("department not found: %w", err)
	}

	updatedDepartment, err := s.repo.UpdateDepartment(id, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update department: %w", err)
	}

	return updatedDepartment, nil
}

func (s *DepartmentService) DeleteDepartment(id string) error {
	if id == "" {
		return errors.New("department ID is required")
	}

	// Verify department exists
	_, err := s.repo.GetDepartment(id)
	if err != nil {
		return fmt.Errorf("department not found: %w", err)
	}

	// Soft delete by setting is_active = false
	isActive := false
	_, err = s.repo.UpdateDepartment(id, &models.UpdateDepartmentRequest{
		IsActive: &isActive,
	})

	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}

	return nil
}