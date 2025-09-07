package service

import (
	"errors"
	"fmt"
	"product-service/internal/models"
	"product-service/internal/repository"
)

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{
		repo: repo,
	}
}

func (s *ProductService) GetProduct(id string) (*models.Product, error) {
	if id == "" {
		return nil, errors.New("product ID is required")
	}

	product, err := s.repo.GetProduct(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (s *ProductService) ListProducts(filter models.ProductFilter) (*models.ProductListResponse, error) {
	// Set default pagination values
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	response, err := s.repo.ListProducts(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	return response, nil
}

func (s *ProductService) CreateProduct(request *models.CreateProductRequest) (*models.Product, error) {
	if request == nil {
		return nil, errors.New("create product request is required")
	}

	// Check if SKU already exists
	// Note: In a real implementation, you'd want to add a GSI on SKU for efficient lookups
	
	product := &models.Product{
		Name:          request.Name,
		Description:   request.Description,
		SKU:           request.SKU,
		Slug:          request.Slug,
		Price:         request.Price,
		OriginalPrice: request.OriginalPrice,
		CategoryID:    request.CategoryID,
		DepartmentID:  request.DepartmentID,
		Brand:         request.Brand,
		Unit:          request.Unit,
		Images:        request.Images,
		Stock:         request.Stock,
		MinStock:      request.MinStock,
		Weight:        request.Weight,
		WeightUnit:    request.WeightUnit,
		Dimensions:    request.Dimensions,
		IsOnSale:      request.IsOnSale,
		Discount:      request.Discount,
		Rating:        0.0, // Default rating for new products
		Reviews:       0,   // Default reviews count for new products
		Tags:          request.Tags,
	}

	err := s.repo.CreateProduct(product)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (s *ProductService) UpdateProduct(id string, request *models.UpdateProductRequest) (*models.Product, error) {
	if id == "" {
		return nil, errors.New("product ID is required")
	}

	if request == nil {
		return nil, errors.New("update product request is required")
	}

	// Verify product exists
	_, err := s.repo.GetProduct(id)
	if err != nil {
		return nil, fmt.Errorf("product not found: %w", err)
	}

	updatedProduct, err := s.repo.UpdateProduct(id, request)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return updatedProduct, nil
}

func (s *ProductService) DeleteProduct(id string) error {
	if id == "" {
		return errors.New("product ID is required")
	}

	// Verify product exists
	_, err := s.repo.GetProduct(id)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Instead of hard delete, we could soft delete by setting is_active = false
	isActive := false
	_, err = s.repo.UpdateProduct(id, &models.UpdateProductRequest{
		IsActive: &isActive,
	})
	
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (s *ProductService) UpdateStock(id string, quantity int) error {
	if id == "" {
		return errors.New("product ID is required")
	}

	// Get current product to check stock constraints
	product, err := s.repo.GetProduct(id)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// Check if the operation would result in negative stock
	newStock := product.Stock + quantity
	if newStock < 0 {
		return fmt.Errorf("insufficient stock: current=%d, requested=%d", product.Stock, quantity)
	}

	err = s.repo.UpdateStock(id, quantity)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

func (s *ProductService) GetLowStockProducts() ([]models.Product, error) {
	products, err := s.repo.GetLowStockProducts()
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	return products, nil
}

func (s *ProductService) SearchProducts(query string, filter models.ProductFilter) (*models.ProductListResponse, error) {
	if query == "" {
		return s.ListProducts(filter)
	}

	filter.Search = query
	return s.ListProducts(filter)
}

func (s *ProductService) GetProductsByCategory(categoryID string, filter models.ProductFilter) (*models.ProductListResponse, error) {
	if categoryID == "" {
		return nil, errors.New("category ID is required")
	}

	filter.CategoryID = categoryID
	return s.ListProducts(filter)
}

func (s *ProductService) GetProductsByBrand(brand string, filter models.ProductFilter) (*models.ProductListResponse, error) {
	if brand == "" {
		return nil, errors.New("brand is required")
	}

	filter.Brand = brand
	return s.ListProducts(filter)
}

func (s *ProductService) GetProductsByDepartment(departmentID string, filter models.ProductFilter) (*models.ProductListResponse, error) {
	if departmentID == "" {
		return nil, errors.New("department ID is required")
	}

	filter.DepartmentID = departmentID
	return s.ListProducts(filter)
}

func (s *ProductService) GetProductsOnSale(filter models.ProductFilter) (*models.ProductListResponse, error) {
	onSale := true
	filter.IsOnSale = &onSale
	return s.ListProducts(filter)
}