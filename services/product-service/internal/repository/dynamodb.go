package repository

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"product-service/internal/models"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type ProductRepository interface {
	GetProduct(id string) (*models.Product, error)
	ListProducts(filter models.ProductFilter) (*models.ProductListResponse, error)
	CreateProduct(product *models.Product) error
	UpdateProduct(id string, updates *models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(id string) error
	UpdateStock(id string, quantity int) error
	GetLowStockProducts() ([]models.Product, error)
	
	// Department methods
	GetDepartment(id string) (*models.Department, error)
	ListDepartments() ([]models.Department, error)
	CreateDepartment(department *models.Department) error
	UpdateDepartment(id string, updates *models.UpdateDepartmentRequest) (*models.Department, error)
	DeleteDepartment(id string) error
	
	// Category methods
	GetCategory(id string) (*models.Category, error)
	ListCategories(parentID *string) ([]models.Category, error)
	CreateCategory(category *models.Category) error
	UpdateCategory(id string, updates *models.UpdateCategoryRequest) (*models.Category, error)
	DeleteCategory(id string) error
}

type DynamoDBRepository struct {
	client    *dynamodb.DynamoDB
	tableName string
}

func NewDynamoDBRepository(tableName string) *DynamoDBRepository {
	sess := session.Must(session.NewSession())
	client := dynamodb.New(sess)
	
	return &DynamoDBRepository{
		client:    client,
		tableName: tableName,
	}
}

func (r *DynamoDBRepository) GetProduct(id string) (*models.Product, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
	}

	result, err := r.client.GetItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if result.Item == nil {
		return nil, errors.New("product not found")
	}

	var product models.Product
	err = dynamodbattribute.UnmarshalMap(result.Item, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal product: %w", err)
	}

	return &product, nil
}

func (r *DynamoDBRepository) ListProducts(filter models.ProductFilter) (*models.ProductListResponse, error) {
	var filterExpression []string
	var expressionAttributeNames map[string]*string
	var expressionAttributeValues map[string]*dynamodb.AttributeValue

	expressionAttributeNames = make(map[string]*string)
	expressionAttributeValues = make(map[string]*dynamodb.AttributeValue)

	// Build filter expression
	if filter.CategoryID != "" {
		filterExpression = append(filterExpression, "#category_id = :category_id")
		expressionAttributeNames["#category_id"] = aws.String("category_id")
		expressionAttributeValues[":category_id"] = &dynamodb.AttributeValue{S: aws.String(filter.CategoryID)}
	}

	if filter.DepartmentID != "" {
		filterExpression = append(filterExpression, "#department_id = :department_id")
		expressionAttributeNames["#department_id"] = aws.String("department_id")
		expressionAttributeValues[":department_id"] = &dynamodb.AttributeValue{S: aws.String(filter.DepartmentID)}
	}

	if filter.Brand != "" {
		filterExpression = append(filterExpression, "#brand = :brand")
		expressionAttributeNames["#brand"] = aws.String("brand")
		expressionAttributeValues[":brand"] = &dynamodb.AttributeValue{S: aws.String(filter.Brand)}
	}

	if filter.MinPrice != nil {
		filterExpression = append(filterExpression, "#price >= :min_price")
		expressionAttributeNames["#price"] = aws.String("price")
		expressionAttributeValues[":min_price"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *filter.MinPrice))}
	}

	if filter.MaxPrice != nil {
		filterExpression = append(filterExpression, "#price <= :max_price")
		expressionAttributeNames["#price"] = aws.String("price")
		expressionAttributeValues[":max_price"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *filter.MaxPrice))}
	}

	if filter.InStock != nil && *filter.InStock {
		filterExpression = append(filterExpression, "#stock > :zero")
		expressionAttributeNames["#stock"] = aws.String("stock")
		expressionAttributeValues[":zero"] = &dynamodb.AttributeValue{N: aws.String("0")}
	}

	if filter.IsOnSale != nil && *filter.IsOnSale {
		filterExpression = append(filterExpression, "#is_on_sale = :is_on_sale")
		expressionAttributeNames["#is_on_sale"] = aws.String("is_on_sale")
		expressionAttributeValues[":is_on_sale"] = &dynamodb.AttributeValue{BOOL: aws.Bool(true)}
	}

	if filter.MinRating != nil {
		filterExpression = append(filterExpression, "#rating >= :min_rating")
		expressionAttributeNames["#rating"] = aws.String("rating")
		expressionAttributeValues[":min_rating"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *filter.MinRating))}
	}

	// Always filter for active products
	filterExpression = append(filterExpression, "#is_active = :is_active")
	expressionAttributeNames["#is_active"] = aws.String("is_active")
	expressionAttributeValues[":is_active"] = &dynamodb.AttributeValue{BOOL: aws.Bool(true)}

	input := &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
	}

	if len(filterExpression) > 0 {
		input.FilterExpression = aws.String(strings.Join(filterExpression, " AND "))
		input.ExpressionAttributeNames = expressionAttributeNames
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	if filter.Limit > 0 {
		input.Limit = aws.Int64(int64(filter.Limit))
	}

	result, err := r.client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to scan products: %w", err)
	}

	var products []models.Product
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	// Apply search filter in memory (for simplicity)
	if filter.Search != "" {
		filteredProducts := make([]models.Product, 0)
		searchLower := strings.ToLower(filter.Search)
		for _, product := range products {
			if strings.Contains(strings.ToLower(product.Name), searchLower) ||
				strings.Contains(strings.ToLower(product.Description), searchLower) ||
				strings.Contains(strings.ToLower(product.SKU), searchLower) ||
				strings.Contains(strings.ToLower(product.Brand), searchLower) ||
				strings.Contains(strings.ToLower(product.Slug), searchLower) {
				filteredProducts = append(filteredProducts, product)
			}
		}
		products = filteredProducts
	}

	// Apply offset
	if filter.Offset > 0 && filter.Offset < len(products) {
		products = products[filter.Offset:]
	} else if filter.Offset >= len(products) {
		products = []models.Product{}
	}

	totalCount := len(products)
	if filter.Limit > 0 && filter.Limit < len(products) {
		products = products[:filter.Limit]
	}

	return &models.ProductListResponse{
		Products:   products,
		TotalCount: totalCount,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
	}, nil
}

func (r *DynamoDBRepository) CreateProduct(product *models.Product) error {
	product.ID = uuid.New().String()
	product.CreatedAt = time.Now().UTC()
	product.UpdatedAt = time.Now().UTC()
	product.IsActive = true

	item, err := dynamodbattribute.MarshalMap(product)
	if err != nil {
		return fmt.Errorf("failed to marshal product: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.client.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

func (r *DynamoDBRepository) UpdateProduct(id string, updates *models.UpdateProductRequest) (*models.Product, error) {
	var updateExpression []string
	var expressionAttributeNames map[string]*string
	var expressionAttributeValues map[string]*dynamodb.AttributeValue

	expressionAttributeNames = make(map[string]*string)
	expressionAttributeValues = make(map[string]*dynamodb.AttributeValue)

	if updates.Name != nil {
		updateExpression = append(updateExpression, "#name = :name")
		expressionAttributeNames["#name"] = aws.String("name")
		expressionAttributeValues[":name"] = &dynamodb.AttributeValue{S: aws.String(*updates.Name)}
	}

	if updates.Description != nil {
		updateExpression = append(updateExpression, "#description = :description")
		expressionAttributeNames["#description"] = aws.String("description")
		expressionAttributeValues[":description"] = &dynamodb.AttributeValue{S: aws.String(*updates.Description)}
	}

	if updates.Price != nil {
		updateExpression = append(updateExpression, "#price = :price")
		expressionAttributeNames["#price"] = aws.String("price")
		expressionAttributeValues[":price"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *updates.Price))}
	}

	if updates.CategoryID != nil {
		updateExpression = append(updateExpression, "#category_id = :category_id")
		expressionAttributeNames["#category_id"] = aws.String("category_id")
		expressionAttributeValues[":category_id"] = &dynamodb.AttributeValue{S: aws.String(*updates.CategoryID)}
	}

	if updates.Slug != nil {
		updateExpression = append(updateExpression, "#slug = :slug")
		expressionAttributeNames["#slug"] = aws.String("slug")
		expressionAttributeValues[":slug"] = &dynamodb.AttributeValue{S: aws.String(*updates.Slug)}
	}

	if updates.OriginalPrice != nil {
		updateExpression = append(updateExpression, "#original_price = :original_price")
		expressionAttributeNames["#original_price"] = aws.String("original_price")
		expressionAttributeValues[":original_price"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *updates.OriginalPrice))}
	}

	if updates.DepartmentID != nil {
		updateExpression = append(updateExpression, "#department_id = :department_id")
		expressionAttributeNames["#department_id"] = aws.String("department_id")
		expressionAttributeValues[":department_id"] = &dynamodb.AttributeValue{S: aws.String(*updates.DepartmentID)}
	}

	if updates.Brand != nil {
		updateExpression = append(updateExpression, "#brand = :brand")
		expressionAttributeNames["#brand"] = aws.String("brand")
		expressionAttributeValues[":brand"] = &dynamodb.AttributeValue{S: aws.String(*updates.Brand)}
	}

	if updates.Unit != nil {
		updateExpression = append(updateExpression, "#unit = :unit")
		expressionAttributeNames["#unit"] = aws.String("unit")
		expressionAttributeValues[":unit"] = &dynamodb.AttributeValue{S: aws.String(*updates.Unit)}
	}

	if updates.Stock != nil {
		updateExpression = append(updateExpression, "#stock = :stock")
		expressionAttributeNames["#stock"] = aws.String("stock")
		expressionAttributeValues[":stock"] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(*updates.Stock))}
	}

	if updates.Weight != nil {
		updateExpression = append(updateExpression, "#weight = :weight")
		expressionAttributeNames["#weight"] = aws.String("weight")
		expressionAttributeValues[":weight"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *updates.Weight))}
	}

	if updates.WeightUnit != nil {
		updateExpression = append(updateExpression, "#weight_unit = :weight_unit")
		expressionAttributeNames["#weight_unit"] = aws.String("weight_unit")
		expressionAttributeValues[":weight_unit"] = &dynamodb.AttributeValue{S: aws.String(*updates.WeightUnit)}
	}

	if updates.IsOnSale != nil {
		updateExpression = append(updateExpression, "#is_on_sale = :is_on_sale")
		expressionAttributeNames["#is_on_sale"] = aws.String("is_on_sale")
		expressionAttributeValues[":is_on_sale"] = &dynamodb.AttributeValue{BOOL: aws.Bool(*updates.IsOnSale)}
	}

	if updates.Discount != nil {
		updateExpression = append(updateExpression, "#discount = :discount")
		expressionAttributeNames["#discount"] = aws.String("discount")
		expressionAttributeValues[":discount"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *updates.Discount))}
	}

	if updates.Rating != nil {
		updateExpression = append(updateExpression, "#rating = :rating")
		expressionAttributeNames["#rating"] = aws.String("rating")
		expressionAttributeValues[":rating"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", *updates.Rating))}
	}

	if updates.Reviews != nil {
		updateExpression = append(updateExpression, "#reviews = :reviews")
		expressionAttributeNames["#reviews"] = aws.String("reviews")
		expressionAttributeValues[":reviews"] = &dynamodb.AttributeValue{N: aws.String(strconv.Itoa(*updates.Reviews))}
	}

	if updates.IsActive != nil {
		updateExpression = append(updateExpression, "#is_active = :is_active")
		expressionAttributeNames["#is_active"] = aws.String("is_active")
		expressionAttributeValues[":is_active"] = &dynamodb.AttributeValue{BOOL: aws.Bool(*updates.IsActive)}
	}

	// Always update the updated_at timestamp
	updateExpression = append(updateExpression, "#updated_at = :updated_at")
	expressionAttributeNames["#updated_at"] = aws.String("updated_at")
	expressionAttributeValues[":updated_at"] = &dynamodb.AttributeValue{S: aws.String(time.Now().UTC().Format(time.RFC3339))}

	if len(updateExpression) == 0 {
		return nil, errors.New("no fields to update")
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		UpdateExpression:          aws.String("SET " + strings.Join(updateExpression, ", ")),
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
		ReturnValues:              aws.String("ALL_NEW"),
	}

	result, err := r.client.UpdateItem(input)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	var product models.Product
	err = dynamodbattribute.UnmarshalMap(result.Attributes, &product)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal updated product: %w", err)
	}

	return &product, nil
}

func (r *DynamoDBRepository) DeleteProduct(id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
	}

	_, err := r.client.DeleteItem(input)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func (r *DynamoDBRepository) UpdateStock(id string, quantity int) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {S: aws.String(id)},
		},
		UpdateExpression: aws.String("SET #stock = #stock + :quantity, #updated_at = :updated_at"),
		ExpressionAttributeNames: map[string]*string{
			"#stock":      aws.String("stock"),
			"#updated_at": aws.String("updated_at"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":quantity":   {N: aws.String(strconv.Itoa(quantity))},
			":updated_at": {S: aws.String(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	_, err := r.client.UpdateItem(input)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

func (r *DynamoDBRepository) GetLowStockProducts() ([]models.Product, error) {
	input := &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("#stock <= #min_stock AND #is_active = :is_active"),
		ExpressionAttributeNames: map[string]*string{
			"#stock":     aws.String("stock"),
			"#min_stock": aws.String("min_stock"),
			"#is_active": aws.String("is_active"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":is_active": {BOOL: aws.Bool(true)},
		},
	}

	result, err := r.client.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	var products []models.Product
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &products)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal products: %w", err)
	}

	return products, nil
}