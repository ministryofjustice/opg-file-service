package dynamo

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// allows us to mock dynamodb.DynamoDB in our tests
type DBClient interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
}
