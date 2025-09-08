package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"product-service/internal/models"
	"product-service/internal/repository"
	"product-service/internal/service"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/go-playground/validator/v10"
)

type LambdaHandler struct {
	productService    *service.ProductService
	departmentService *service.DepartmentService
	categoryService   *service.CategoryService
	validator         *validator.Validate
}

func NewLambdaHandler() *LambdaHandler {
	repo, err := repository.NewPostgresRepository()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	productService := service.NewProductService(repo)
	departmentService := service.NewDepartmentService(repo)
	categoryService := service.NewCategoryService(repo)
	validator := validator.New()

	return &LambdaHandler{
		productService:    productService,
		departmentService: departmentService,
		categoryService:   categoryService,
		validator:         validator,
	}
}

func (h *LambdaHandler) HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Enable CORS
	headers := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
	}

	// Handle preflight OPTIONS request
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers:    headers,
		}, nil
	}

	switch {
	// Product routes
	case request.HTTPMethod == "GET" && request.Path == "/product-service/products":
		return h.listProducts(request, headers)
	case request.HTTPMethod == "GET" && strings.HasPrefix(request.Path, "/product-service/products/"):
		return h.getProduct(request, headers)
	case request.HTTPMethod == "POST" && request.Path == "/product-service/products":
		return h.createProduct(request, headers)
	case request.HTTPMethod == "PUT" && strings.HasPrefix(request.Path, "/product-service/products/"):
		return h.updateProduct(request, headers)
	case request.HTTPMethod == "DELETE" && strings.HasPrefix(request.Path, "/product-service/products/"):
		return h.deleteProduct(request, headers)
	case request.HTTPMethod == "GET" && request.Path == "/product-service/products/low-stock":
		return h.getLowStockProducts(request, headers)
	case request.HTTPMethod == "POST" && strings.HasPrefix(request.Path, "/product-service/products/") && strings.HasSuffix(request.Path, "/stock"):
		return h.updateStock(request, headers)
	case request.HTTPMethod == "GET" && request.Path == "/product-service/products/on-sale":
		return h.getProductsOnSale(request, headers)
	case request.HTTPMethod == "GET" && strings.HasPrefix(request.Path, "/product-service/products/department/"):
		return h.getProductsByDepartment(request, headers)
	
	// Department routes
	case request.HTTPMethod == "GET" && request.Path == "/product-service/departments":
		return h.listDepartments(request, headers)
	case request.HTTPMethod == "GET" && strings.HasPrefix(request.Path, "/product-service/departments/"):
		return h.getDepartment(request, headers)
	case request.HTTPMethod == "POST" && request.Path == "/product-service/departments":
		return h.createDepartment(request, headers)
	case request.HTTPMethod == "PUT" && strings.HasPrefix(request.Path, "/product-service/departments/"):
		return h.updateDepartment(request, headers)
	case request.HTTPMethod == "DELETE" && strings.HasPrefix(request.Path, "/product-service/departments/"):
		return h.deleteDepartment(request, headers)
	
	// Category routes
	case request.HTTPMethod == "GET" && request.Path == "/product-service/categories":
		return h.listCategories(request, headers)
	case request.HTTPMethod == "GET" && strings.HasPrefix(request.Path, "/product-service/categories/"):
		return h.getCategory(request, headers)
	case request.HTTPMethod == "POST" && request.Path == "/product-service/categories":
		return h.createCategory(request, headers)
	case request.HTTPMethod == "PUT" && strings.HasPrefix(request.Path, "/product-service/categories/"):
		return h.updateCategory(request, headers)
	case request.HTTPMethod == "DELETE" && strings.HasPrefix(request.Path, "/product-service/categories/"):
		return h.deleteCategory(request, headers)
	
	default:
		return h.errorResponse(http.StatusNotFound, "Route not found", headers), nil
	}
}

func (h *LambdaHandler) listProducts(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	filter := models.ProductFilter{}

	// Parse query parameters
	if categoryID := request.QueryStringParameters["category_id"]; categoryID != "" {
		filter.CategoryID = categoryID
	}
	if brand := request.QueryStringParameters["brand"]; brand != "" {
		filter.Brand = brand
	}
	if search := request.QueryStringParameters["search"]; search != "" {
		filter.Search = search
	}
	if minPriceStr := request.QueryStringParameters["min_price"]; minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filter.MinPrice = &minPrice
		}
	}
	if maxPriceStr := request.QueryStringParameters["max_price"]; maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filter.MaxPrice = &maxPrice
		}
	}
	if inStockStr := request.QueryStringParameters["in_stock"]; inStockStr != "" {
		if inStock, err := strconv.ParseBool(inStockStr); err == nil {
			filter.InStock = &inStock
		}
	}
	if departmentID := request.QueryStringParameters["department_id"]; departmentID != "" {
		filter.DepartmentID = departmentID
	}
	if isOnSaleStr := request.QueryStringParameters["is_on_sale"]; isOnSaleStr != "" {
		if isOnSale, err := strconv.ParseBool(isOnSaleStr); err == nil {
			filter.IsOnSale = &isOnSale
		}
	}
	if minRatingStr := request.QueryStringParameters["min_rating"]; minRatingStr != "" {
		if minRating, err := strconv.ParseFloat(minRatingStr, 64); err == nil {
			filter.MinRating = &minRating
		}
	}
	if limitStr := request.QueryStringParameters["limit"]; limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := request.QueryStringParameters["offset"]; offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	response, err := h.productService.ListProducts(filter)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, response, headers), nil
}

func (h *LambdaHandler) getProduct(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Product ID is required", headers), nil
	}

	product, err := h.productService.GetProduct(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Product not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, product, headers), nil
}

func (h *LambdaHandler) createProduct(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	var createRequest models.CreateProductRequest

	if err := json.Unmarshal([]byte(request.Body), &createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	product, err := h.productService.CreateProduct(&createRequest)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusCreated, product, headers), nil
}

func (h *LambdaHandler) updateProduct(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Product ID is required", headers), nil
	}

	var updateRequest models.UpdateProductRequest

	if err := json.Unmarshal([]byte(request.Body), &updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	product, err := h.productService.UpdateProduct(id, &updateRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Product not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, product, headers), nil
}

func (h *LambdaHandler) deleteProduct(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Product ID is required", headers), nil
	}

	err := h.productService.DeleteProduct(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Product not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusNoContent, nil, headers), nil
}

func (h *LambdaHandler) getLowStockProducts(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	products, err := h.productService.GetLowStockProducts()
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, products, headers), nil
}

func (h *LambdaHandler) updateStock(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(strings.Replace(request.Path, "/stock", "", 1))
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Product ID is required", headers), nil
	}

	var stockRequest struct {
		Quantity int `json:"quantity" validate:"required"`
	}

	if err := json.Unmarshal([]byte(request.Body), &stockRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&stockRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	err := h.productService.UpdateStock(id, stockRequest.Quantity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Product not found", headers), nil
		}
		if strings.Contains(err.Error(), "insufficient stock") {
			return h.errorResponse(http.StatusBadRequest, err.Error(), headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, map[string]string{"message": "Stock updated successfully"}, headers), nil
}

func (h *LambdaHandler) getProductsOnSale(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	filter := models.ProductFilter{}

	// Parse query parameters for additional filtering
	if limitStr := request.QueryStringParameters["limit"]; limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := request.QueryStringParameters["offset"]; offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	response, err := h.productService.GetProductsOnSale(filter)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, response, headers), nil
}

func (h *LambdaHandler) getProductsByDepartment(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	departmentID := extractDepartmentIDFromPath(request.Path)
	if departmentID == "" {
		return h.errorResponse(http.StatusBadRequest, "Department ID is required", headers), nil
	}

	filter := models.ProductFilter{}

	// Parse query parameters for additional filtering
	if limitStr := request.QueryStringParameters["limit"]; limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}
	if offsetStr := request.QueryStringParameters["offset"]; offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = offset
		}
	}

	response, err := h.productService.GetProductsByDepartment(departmentID, filter)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, response, headers), nil
}

func (h *LambdaHandler) successResponse(statusCode int, data interface{}, headers map[string]string) events.APIGatewayProxyResponse {
	var body string
	if data != nil {
		bodyBytes, _ := json.Marshal(data)
		body = string(bodyBytes)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
}

func (h *LambdaHandler) errorResponse(statusCode int, message string, headers map[string]string) events.APIGatewayProxyResponse {
	errorBody := map[string]string{"error": message}
	bodyBytes, _ := json.Marshal(errorBody)

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(bodyBytes),
	}
}

func extractIDFromPath(path string) string {
	// Remove leading and trailing slashes and split by "/"
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// For paths like "/product-service/products/123" or "/product-service/departments/456"
	// We want to extract the last part which is the ID
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractDepartmentIDFromPath(path string) string {
	// Path format: /product-service/products/department/{departmentID}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[2] == "department" {
		return parts[3]
	}
	return ""
}

// Department handlers

func (h *LambdaHandler) listDepartments(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	departments, err := h.departmentService.ListDepartments()
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, departments, headers), nil
}

func (h *LambdaHandler) getDepartment(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Department ID is required", headers), nil
	}

	department, err := h.departmentService.GetDepartment(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Department not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, department, headers), nil
}

func (h *LambdaHandler) createDepartment(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	var createRequest models.CreateDepartmentRequest

	if err := json.Unmarshal([]byte(request.Body), &createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	department, err := h.departmentService.CreateDepartment(&createRequest)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusCreated, department, headers), nil
}

func (h *LambdaHandler) updateDepartment(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Department ID is required", headers), nil
	}

	var updateRequest models.UpdateDepartmentRequest

	if err := json.Unmarshal([]byte(request.Body), &updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	department, err := h.departmentService.UpdateDepartment(id, &updateRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Department not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, department, headers), nil
}

func (h *LambdaHandler) deleteDepartment(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Department ID is required", headers), nil
	}

	err := h.departmentService.DeleteDepartment(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Department not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusNoContent, nil, headers), nil
}

// Category handlers

func (h *LambdaHandler) listCategories(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	var parentID *string
	if parentIDParam := request.QueryStringParameters["parent_id"]; parentIDParam != "" {
		parentID = &parentIDParam
	}

	categories, err := h.categoryService.ListCategories(parentID)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, categories, headers), nil
}

func (h *LambdaHandler) getCategory(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Category ID is required", headers), nil
	}

	category, err := h.categoryService.GetCategory(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Category not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, category, headers), nil
}

func (h *LambdaHandler) createCategory(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	var createRequest models.CreateCategoryRequest

	if err := json.Unmarshal([]byte(request.Body), &createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&createRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	category, err := h.categoryService.CreateCategory(&createRequest)
	if err != nil {
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusCreated, category, headers), nil
}

func (h *LambdaHandler) updateCategory(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Category ID is required", headers), nil
	}

	var updateRequest models.UpdateCategoryRequest

	if err := json.Unmarshal([]byte(request.Body), &updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, "Invalid JSON payload", headers), nil
	}

	if err := h.validator.Struct(&updateRequest); err != nil {
		return h.errorResponse(http.StatusBadRequest, fmt.Sprintf("Validation error: %s", err.Error()), headers), nil
	}

	category, err := h.categoryService.UpdateCategory(id, &updateRequest)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Category not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusOK, category, headers), nil
}

func (h *LambdaHandler) deleteCategory(request events.APIGatewayProxyRequest, headers map[string]string) (events.APIGatewayProxyResponse, error) {
	id := extractIDFromPath(request.Path)
	if id == "" {
		return h.errorResponse(http.StatusBadRequest, "Category ID is required", headers), nil
	}

	err := h.categoryService.DeleteCategory(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return h.errorResponse(http.StatusNotFound, "Category not found", headers), nil
		}
		return h.errorResponse(http.StatusInternalServerError, err.Error(), headers), nil
	}

	return h.successResponse(http.StatusNoContent, nil, headers), nil
}
