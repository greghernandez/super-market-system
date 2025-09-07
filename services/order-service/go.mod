module order-service

go 1.21

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/go-playground/validator/v10 v10.16.0
	github.com/google/uuid v1.4.0
	shared/db v0.0.0
)

replace shared/db => ../../shared/db