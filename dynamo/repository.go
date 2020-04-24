package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
	"opg-file-service/internal"
	"opg-file-service/session"
	"opg-file-service/storage"
	"os"
)

type Repository struct {
	db     *dynamodb.DynamoDB
	logger *log.Logger
	table  string
}

func NewRepository(sess session.Session, l *log.Logger) *Repository {
	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	sess.AwsSession.Config.Endpoint = &endpoint

	dynamo := dynamodb.New(sess.AwsSession)

	return &Repository{
		db:     dynamo,
		logger: l,
		table:  internal.GetEnvVar("AWS_DYNAMODB_TABLE_NAME", "zip-requests"),
	}
}

func (repo Repository) Get(ref string) (*storage.Entry, error) {
	notFound := storage.NotFoundError{Ref: ref}

	result, err := repo.db.GetItem(&dynamodb.GetItemInput{
		TableName: &repo.table,
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

func (repo Repository) Delete(entry *storage.Entry) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &repo.table,
		Key: map[string]*dynamodb.AttributeValue{
			"ref": {
				S: aws.String(entry.Ref),
			},
		},
	}

	_, err := repo.db.DeleteItem(input)

	return err
}
