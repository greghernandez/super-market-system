package models

import (
	"time"
)

type ProductDimensions struct {
	Length float64 `json:"length" gorm:"type:decimal(10,2)" validate:"min=0"` // cm
	Width  float64 `json:"width" gorm:"type:decimal(10,2)" validate:"min=0"`   // cm
	Height float64 `json:"height" gorm:"type:decimal(10,2)" validate:"min=0"` // cm
}

type Product struct {
	ID            string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SKU           string            `json:"sku" gorm:"uniqueIndex;not null" validate:"required"`
	Slug          string            `json:"slug" gorm:"uniqueIndex;not null" validate:"required"`
	Name          string            `json:"name" gorm:"not null" validate:"required,min=1,max=255"`
	Description   string            `json:"description" gorm:"type:text"`
	Price         float64           `json:"price" gorm:"type:decimal(10,2);not null" validate:"required,min=0"`
	OriginalPrice *float64          `json:"original_price" gorm:"type:decimal(10,2)" validate:"omitempty,min=0"`
	Images        []string          `json:"images" gorm:"type:text[]"`
	Category      *Category         `json:"category,omitempty" gorm:"foreignKey:CategoryID"`
	CategoryID    string            `json:"category_id" gorm:"type:uuid;not null;index" validate:"required"`
	DepartmentID  string            `json:"department_id" gorm:"type:uuid;not null;index" validate:"required"`
	Brand         string            `json:"brand" gorm:"index"`
	Unit          string            `json:"unit"`
	Stock         int               `json:"stock" gorm:"not null;default:0" validate:"min=0"`
	MinStock      int               `json:"min_stock" gorm:"not null;default:0" validate:"min=0"`
	Weight        float64           `json:"weight" gorm:"type:decimal(10,3)" validate:"min=0"` // gramos
	WeightUnit    string            `json:"weight_unit"`
	Dimensions    ProductDimensions `json:"dimensions" gorm:"embedded;embeddedPrefix:dim_"`
	IsOnSale      bool              `json:"is_on_sale" gorm:"index;default:false"`
	Discount      *float64          `json:"discount" gorm:"type:decimal(5,2)" validate:"omitempty,min=0,max=100"`
	Rating        float64           `json:"rating" gorm:"type:decimal(3,2);default:0" validate:"min=0,max=5"`
	Reviews       int               `json:"reviews" gorm:"default:0" validate:"min=0"`
	IsActive      bool              `json:"is_active" gorm:"index;default:true"`
	Tags          []string          `json:"tags" gorm:"type:text[]"`
	CreatedAt     time.Time         `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time         `json:"updated_at" gorm:"autoUpdateTime"`
}

type Department struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null" validate:"required,min=1,max=100"`
	Description string    `json:"description" gorm:"type:text"`
	Icon        string    `json:"icon"`
	Image       string    `json:"image"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null" validate:"required"`
	IsActive    bool      `json:"is_active" gorm:"index;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Category struct {
	ID          string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"not null" validate:"required,min=1,max=100"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null" validate:"required"`
	Description string    `json:"description" gorm:"type:text"`
	ParentID    *string   `json:"parent_id" gorm:"type:uuid;index"`
	Level       int       `json:"level" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"index;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ProductFilter struct {
	CategoryID   string   `json:"category_id"`
	DepartmentID string   `json:"department_id"`
	Brand        string   `json:"brand"`
	MinPrice     *float64 `json:"min_price"`
	MaxPrice     *float64 `json:"max_price"`
	InStock      *bool    `json:"in_stock"`
	IsOnSale     *bool    `json:"is_on_sale"`
	MinRating    *float64 `json:"min_rating"`
	Search       string   `json:"search"`
	Tags         []string `json:"tags"`
	Limit        int      `json:"limit" validate:"min=1,max=100"`
	Offset       int      `json:"offset" validate:"min=0"`
}

type ProductListResponse struct {
	Products   []Product `json:"products"`
	TotalCount int       `json:"total_count"`
	Limit      int       `json:"limit"`
	Offset     int       `json:"offset"`
}

type CreateProductRequest struct {
	Name          string            `json:"name" validate:"required,min=1,max=255"`
	Description   string            `json:"description"`
	SKU           string            `json:"sku" validate:"required"`
	Slug          string            `json:"slug" validate:"required"`
	Price         float64           `json:"price" validate:"required,min=0"`
	OriginalPrice *float64          `json:"original_price" validate:"omitempty,min=0"`
	CategoryID    string            `json:"category_id" validate:"required"`
	DepartmentID  string            `json:"department_id" validate:"required"`
	Brand         string            `json:"brand"`
	Unit          string            `json:"unit"`
	Images        []string          `json:"images"`
	Stock         int               `json:"stock" validate:"min=0"`
	MinStock      int               `json:"min_stock" validate:"min=0"`
	Weight        float64           `json:"weight" validate:"min=0"`
	WeightUnit    string            `json:"weight_unit"`
	Dimensions    ProductDimensions `json:"dimensions"`
	IsOnSale      bool              `json:"is_on_sale"`
	Discount      *float64          `json:"discount" validate:"omitempty,min=0,max=100"`
	Tags          []string          `json:"tags"`
}

type UpdateProductRequest struct {
	Name          *string            `json:"name" validate:"omitempty,min=1,max=255"`
	Description   *string            `json:"description"`
	Slug          *string            `json:"slug"`
	Price         *float64           `json:"price" validate:"omitempty,min=0"`
	OriginalPrice *float64           `json:"original_price" validate:"omitempty,min=0"`
	CategoryID    *string            `json:"category_id"`
	DepartmentID  *string            `json:"department_id"`
	Brand         *string            `json:"brand"`
	Unit          *string            `json:"unit"`
	Images        []string           `json:"images"`
	Stock         *int               `json:"stock" validate:"omitempty,min=0"`
	MinStock      *int               `json:"min_stock" validate:"omitempty,min=0"`
	Weight        *float64           `json:"weight" validate:"omitempty,min=0"`
	WeightUnit    *string            `json:"weight_unit"`
	Dimensions    *ProductDimensions `json:"dimensions"`
	IsOnSale      *bool              `json:"is_on_sale"`
	Discount      *float64           `json:"discount" validate:"omitempty,min=0,max=100"`
	Rating        *float64           `json:"rating" validate:"omitempty,min=0,max=5"`
	Reviews       *int               `json:"reviews" validate:"omitempty,min=0"`
	IsActive      *bool              `json:"is_active"`
	Tags          []string           `json:"tags"`
}

type CreateDepartmentRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Image       string `json:"image"`
	Slug        string `json:"slug" validate:"required"`
}

type UpdateDepartmentRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	Image       *string `json:"image"`
	Slug        *string `json:"slug"`
	IsActive    *bool   `json:"is_active"`
}

type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Slug        string  `json:"slug" validate:"required"`
	Description string  `json:"description"`
	ParentID    *string `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id"`
	IsActive    *bool   `json:"is_active"`
}