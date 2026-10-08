package infrastructure

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/sony/gobreaker"

	"github.com/hsinyuko1111/gospital/internal/middleware"
)

// DynamoDBClient wraps DynamoDB client with circuit breaker
type DynamoDBClient struct {
	client *dynamodb.Client
	cb     *gobreaker.CircuitBreaker
}

// NewDynamoDBClient creates a new DynamoDB client with circuit breaker
func NewDynamoDBClient(client *dynamodb.Client) *DynamoDBClient {
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "DynamoDB",
		MaxRequests: 3,                // Allow 3 requests in half-open state
		Interval:    10 * time.Second, // Reset failure count every 10 seconds
		Timeout:     30 * time.Second, // Stay open for 30 seconds before half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Open circuit if failure ratio >= 60% with at least 3 requests
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			log.Printf("Circuit Breaker '%s' changed from '%s' to '%s'\n", name, from.String(), to.String())
			// Update Prometheus metric
			middleware.UpdateDynamoDBCircuitBreakerState(to.String())
		},
	})

	// Initialize metric to closed state
	middleware.UpdateDynamoDBCircuitBreakerState("closed")

	return &DynamoDBClient{
		client: client,
		cb:     cb,
	}
}

// PutItem wraps DynamoDB PutItem with circuit breaker
func (c *DynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.PutItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.PutItemOutput), nil
}

// GetItem wraps DynamoDB GetItem with circuit breaker
func (c *DynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.GetItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.GetItemOutput), nil
}

// Query wraps DynamoDB Query with circuit breaker
func (c *DynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.Query(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.QueryOutput), nil
}

// UpdateItem wraps DynamoDB UpdateItem with circuit breaker
func (c *DynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.UpdateItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.UpdateItemOutput), nil
}

// DeleteItem wraps DynamoDB DeleteItem with circuit breaker
func (c *DynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.DeleteItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.DeleteItemOutput), nil
}

// Scan wraps DynamoDB Scan with circuit breaker
func (c *DynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.Scan(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.ScanOutput), nil
}

// BatchGetItem wraps DynamoDB BatchGetItem with circuit breaker
func (c *DynamoDBClient) BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.BatchGetItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.BatchGetItemOutput), nil
}

// BatchWriteItem wraps DynamoDB BatchWriteItem with circuit breaker
func (c *DynamoDBClient) BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.client.BatchWriteItem(ctx, params, optFns...)
	})
	if err != nil {
		return nil, err
	}
	return result.(*dynamodb.BatchWriteItemOutput), nil
}

// GetClient returns the underlying DynamoDB client for operations not wrapped
func (c *DynamoDBClient) GetClient() *dynamodb.Client {
	return c.client
}
