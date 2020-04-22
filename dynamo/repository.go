package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"opg-s3-zipper-service/session"
	"opg-s3-zipper-service/storage"
	"os"
)

type Repository struct {
	db     *dynamodb.DynamoDB
	logger *log.Logger
}

func NewRepository(sess session.Session, l *log.Logger) *Repository {
	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	sess.AwsSession.Config.Endpoint = &endpoint

	dynamo := dynamodb.New(sess.AwsSession)

	return &Repository{
		db:     dynamo,
		logger: l,
	}
}

func (repo Repository) GetEntry(ref string) (*storage.Entry, error) {
	tableName := os.Getenv("AWS_DYNAMODB_TABLE_NAME")

	notFound := storage.NotFoundError{Ref: ref}

	result, err := repo.db.GetItem(&dynamodb.GetItemInput{
		TableName: &tableName,
		Key: map[string]*dynamodb.AttributeValue{
			"ref": {
				S: aws.String(ref),
			},
		},
	})
	if err != nil {
		repo.logger.Println(err.Error())
		return nil, notFound
	}

	entry := storage.Entry{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &entry)
	if err != nil {
		repo.logger.Println("Failed to unmarshal Record, ", err)
		return nil, notFound
	}

	if entry.Ref == "" {
		repo.logger.Println("Ref token " + ref + " has expired or does not exist.")
		return nil, notFound
	}

	return &entry, nil
}
