package repository

import (
	"errors"
	"fmt"
	"strings"

	"product-service/internal/models"
	"shared/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresRepository struct {
	*db.BaseRepository
}

func NewPostgresRepository() (*PostgresRepository, error) {
	database, err := db.NewPostgresConnectionFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// // Auto-migrate the schema
	// err = db.AutoMigrate(database, &models.Product{}, &models.Category{})
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to migrate database: %w", err)
	// }

	return &PostgresRepository{
		BaseRepository: db.NewBaseRepository(database),
	}, nil
}

func (r *PostgresRepository) GetProduct(id string) (*models.Product, error) {
	var product models.Product
	result := r.DB.Where("id = ? AND is_active = ?", id, true).First(&product)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("product not found")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get product: %w", result.Error)
	}

	return &product, nil
}

func (r *PostgresRepository) ListProducts(filter models.ProductFilter) (*models.ProductListResponse, error) {
	query := r.DB.Model(&models.Product{}).Where("is_active = ?", true)

	// Apply filters
	if filter.CategoryID != "" {
		query = query.Where("category_id = ?", filter.CategoryID)
	}

	if filter.DepartmentID != "" {
		query = query.Where("department_id = ?", filter.DepartmentID)
	}

	if filter.Brand != "" {
		query = query.Where("brand = ?", filter.Brand)
	}

	if filter.MinPrice != nil {
		query = query.Where("price >= ?", *filter.MinPrice)
	}

	if filter.MaxPrice != nil {
		query = query.Where("price <= ?", *filter.MaxPrice)
	}

	if filter.InStock != nil && *filter.InStock {
		query = query.Where("stock > ?", 0)
	}

	if filter.IsOnSale != nil && *filter.IsOnSale {
		query = query.Where("is_on_sale = ?", true)
	}

	if filter.MinRating != nil {
		query = query.Where("rating >= ?", *filter.MinRating)
	}

	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(sku) LIKE ? OR LOWER(brand) LIKE ? OR LOWER(slug) LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Count total records
	var totalCount int64
	query.Count(&totalCount)

	// Apply pagination
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	var products []models.Product
	result := query.Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list products: %w", result.Error)
	}

	return &models.ProductListResponse{
		Products:   products,
		TotalCount: int(totalCount),
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}, nil
}

func (r *PostgresRepository) CreateProduct(product *models.Product) error {
	product.ID = uuid.New().String()
	product.IsActive = true

	result := r.DB.Create(product)
	if result.Error != nil {
		return fmt.Errorf("failed to create product: %w", result.Error)
	}

	return nil
}

func (r *PostgresRepository) UpdateProduct(id string, updates *models.UpdateProductRequest) (*models.Product, error) {
	var product models.Product

	// First get the existing product
	result := r.DB.Where("id = ?", id).First(&product)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("product not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find product: %w", result.Error)
	}

	// Apply updates
	updateFields := make(map[string]interface{})

	if updates.Name != nil {
		updateFields["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateFields["description"] = *updates.Description
	}
	if updates.Price != nil {
		updateFields["price"] = *updates.Price
	}
	if updates.CategoryID != nil {
		updateFields["category_id"] = *updates.CategoryID
	}
	if updates.Slug != nil {
		updateFields["slug"] = *updates.Slug
	}
	if updates.OriginalPrice != nil {
		updateFields["original_price"] = *updates.OriginalPrice
	}
	if updates.DepartmentID != nil {
		updateFields["department_id"] = *updates.DepartmentID
	}
	if updates.Brand != nil {
		updateFields["brand"] = *updates.Brand
	}
	if updates.Unit != nil {
		updateFields["unit"] = *updates.Unit
	}
	if updates.Stock != nil {
		updateFields["stock"] = *updates.Stock
	}
	if updates.Weight != nil {
		updateFields["weight"] = *updates.Weight
	}
	if updates.WeightUnit != nil {
		updateFields["weight_unit"] = *updates.WeightUnit
	}
	if updates.IsOnSale != nil {
		updateFields["is_on_sale"] = *updates.IsOnSale
	}
	if updates.Discount != nil {
		updateFields["discount"] = *updates.Discount
	}
	if updates.Rating != nil {
		updateFields["rating"] = *updates.Rating
	}
	if updates.Reviews != nil {
		updateFields["reviews"] = *updates.Reviews
	}
	if updates.IsActive != nil {
		updateFields["is_active"] = *updates.IsActive
	}

	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	result = r.DB.Model(&product).Updates(updateFields)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update product: %w", result.Error)
	}

	// Return updated product
	result = r.DB.Where("id = ?", id).First(&product)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get updated product: %w", result.Error)
	}

	return &product, nil
}

func (r *PostgresRepository) DeleteProduct(id string) error {
	result := r.DB.Where("id = ?", id).Delete(&models.Product{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete product: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *PostgresRepository) UpdateStock(id string, quantity int) error {
	result := r.DB.Model(&models.Product{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", quantity))

	if result.Error != nil {
		return fmt.Errorf("failed to update stock: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}

	return nil
}

func (r *PostgresRepository) GetLowStockProducts() ([]models.Product, error) {
	var products []models.Product

	result := r.DB.Where("stock <= min_stock AND is_active = ?", true).Find(&products)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", result.Error)
	}

	return products, nil
}

// Department methods

func (r *PostgresRepository) GetDepartment(id string) (*models.Department, error) {
	var department models.Department
	result := r.DB.Where("id = ? AND is_active = ?", id, true).First(&department)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("department not found")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get department: %w", result.Error)
	}

	return &department, nil
}

func (r *PostgresRepository) ListDepartments() ([]models.Department, error) {
	var departments []models.Department
	result := r.DB.Where("is_active = ?", true).Find(&departments)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list departments: %w", result.Error)
	}

	return departments, nil
}

func (r *PostgresRepository) CreateDepartment(department *models.Department) error {
	department.ID = uuid.New().String()
	department.IsActive = true

	result := r.DB.Create(department)
	if result.Error != nil {
		return fmt.Errorf("failed to create department: %w", result.Error)
	}

	return nil
}

func (r *PostgresRepository) UpdateDepartment(id string, updates *models.UpdateDepartmentRequest) (*models.Department, error) {
	var department models.Department

	// First get the existing department
	result := r.DB.Where("id = ?", id).First(&department)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("department not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find department: %w", result.Error)
	}

	// Apply updates
	updateFields := make(map[string]interface{})

	if updates.Name != nil {
		updateFields["name"] = *updates.Name
	}
	if updates.Description != nil {
		updateFields["description"] = *updates.Description
	}
	if updates.Icon != nil {
		updateFields["icon"] = *updates.Icon
	}
	if updates.Image != nil {
		updateFields["image"] = *updates.Image
	}
	if updates.Slug != nil {
		updateFields["slug"] = *updates.Slug
	}
	if updates.IsActive != nil {
		updateFields["is_active"] = *updates.IsActive
	}

	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	result = r.DB.Model(&department).Updates(updateFields)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update department: %w", result.Error)
	}

	// Return updated department
	result = r.DB.Where("id = ?", id).First(&department)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get updated department: %w", result.Error)
	}

	return &department, nil
}

func (r *PostgresRepository) DeleteDepartment(id string) error {
	// Soft delete by setting is_active = false
	result := r.DB.Model(&models.Department{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to delete department: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("department not found")
	}

	return nil
}

// Category methods

func (r *PostgresRepository) GetCategory(id string) (*models.Category, error) {
	var category models.Category
	result := r.DB.Where("id = ? AND is_active = ?", id, true).First(&category)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("category not found")
	}

	if result.Error != nil {
		return nil, fmt.Errorf("failed to get category: %w", result.Error)
	}

	return &category, nil
}

func (r *PostgresRepository) ListCategories(parentID *string) ([]models.Category, error) {
	query := r.DB.Where("is_active = ?", true)
	
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}

	var categories []models.Category
	result := query.Find(&categories)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list categories: %w", result.Error)
	}

	return categories, nil
}

func (r *PostgresRepository) CreateCategory(category *models.Category) error {
	category.ID = uuid.New().String()
	category.IsActive = true

	// Set level based on parent
	if category.ParentID != nil {
		var parent models.Category
		result := r.DB.Where("id = ?", *category.ParentID).First(&parent)
		if result.Error != nil {
			return fmt.Errorf("parent category not found: %w", result.Error)
		}
		category.Level = parent.Level + 1
	} else {
		category.Level = 0
	}

	result := r.DB.Create(category)
	if result.Error != nil {
		return fmt.Errorf("failed to create category: %w", result.Error)
	}

	return nil
}

func (r *PostgresRepository) UpdateCategory(id string, updates *models.UpdateCategoryRequest) (*models.Category, error) {
	var category models.Category

	// First get the existing category
	result := r.DB.Where("id = ?", id).First(&category)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("category not found")
	}
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find category: %w", result.Error)
	}

	// Apply updates
	updateFields := make(map[string]interface{})

	if updates.Name != nil {
		updateFields["name"] = *updates.Name
	}
	if updates.Slug != nil {
		updateFields["slug"] = *updates.Slug
	}
	if updates.Description != nil {
		updateFields["description"] = *updates.Description
	}
	if updates.ParentID != nil {
		updateFields["parent_id"] = *updates.ParentID
		// Update level if parent changes
		if *updates.ParentID != "" {
			var parent models.Category
			result := r.DB.Where("id = ?", *updates.ParentID).First(&parent)
			if result.Error != nil {
				return nil, fmt.Errorf("parent category not found: %w", result.Error)
			}
			updateFields["level"] = parent.Level + 1
		} else {
			updateFields["level"] = 0
		}
	}
	if updates.IsActive != nil {
		updateFields["is_active"] = *updates.IsActive
	}

	if len(updateFields) == 0 {
		return nil, errors.New("no fields to update")
	}

	result = r.DB.Model(&category).Updates(updateFields)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update category: %w", result.Error)
	}

	// Return updated category
	result = r.DB.Where("id = ?", id).First(&category)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get updated category: %w", result.Error)
	}

	return &category, nil
}

func (r *PostgresRepository) DeleteCategory(id string) error {
	// Soft delete by setting is_active = false
	result := r.DB.Model(&models.Category{}).Where("id = ?", id).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to delete category: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("category not found")
	}

	return nil
}
