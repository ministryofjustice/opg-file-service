package dynamo

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/ministryofjustice/opg-file-service/session"
	"github.com/ministryofjustice/opg-file-service/storage"
)

var errNilEntry = errors.New("entry cannot be nil")

type DBClient interface {
	GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error)
	DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error)
	PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error)
}

type Repository struct {
	db     DBClient
	logger *log.Logger
	table  string
}

func NewRepository(sess session.Session, l *log.Logger, endpoint, table string) *Repository {
	sess.AwsSession.Config.Endpoint = &endpoint

	dynamo := dynamodb.New(sess.AwsSession)

	return &Repository{
		db:     dynamo,
		logger: l,
		table:  table,
	}
}

func (repo *Repository) Get(ref string) (*storage.Entry, error) {
	result, err := repo.db.GetItem(&dynamodb.GetItemInput{
		TableName: &repo.table,
		Key: map[string]*dynamodb.AttributeValue{
			"Ref": {
				S: aws.String(ref),
			},
		},
	})
	if err != nil {
		repo.logger.Println(err.Error())
		return nil, fmt.Errorf("could not find entry '%s': %w", ref, err)
	}

	entry := &storage.Entry{}

	err = dynamodbattribute.UnmarshalMap(result.Item, entry)
	if err != nil {
		repo.logger.Println("Failed to unmarshal Record, ", err)
		return nil, fmt.Errorf("could not find entry '%s': %w", ref, err)
	}

	if entry.Ref == "" {
		repo.logger.Println("Ref token " + ref + " has expired or does not exist.")
		return nil, fmt.Errorf("could not find entry '%s'", ref)
	}

	return entry, nil
}

func (repo *Repository) Delete(entry *storage.Entry) error {
	if entry == nil {
		return errNilEntry
	}

	input := &dynamodb.DeleteItemInput{
		TableName: &repo.table,
		Key: map[string]*dynamodb.AttributeValue{
			"Ref": {
				S: aws.String(entry.Ref),
			},
		},
	}

	_, err := repo.db.DeleteItem(input)
	return err
}

func (repo *Repository) Add(entry *storage.Entry) error {
	if entry == nil {
		return errNilEntry
	}

	av, err := dynamodbattribute.MarshalMap(entry)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: &repo.table,
		Item:      av,
	}

	_, err = repo.db.PutItem(input)
	return err
}
