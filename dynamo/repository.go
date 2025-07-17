package dynamo

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log/slog"
	"opg-file-service/internal"
	"opg-file-service/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type RepositoryInterface interface {
	Get(ctx context.Context, ref string) (*storage.Entry, error)
	Delete(ctx context.Context, entry *storage.Entry) error
	Add(ctx context.Context, entry *storage.Entry) error
}

type Repository struct {
	db     DBClient
	logger *slog.Logger
	table  string
}

func NewRepository(cfg *aws.Config, logger *slog.Logger) RepositoryInterface {
	dynamo := dynamodb.NewFromConfig(*cfg)

	return &Repository{
		db:     dynamo,
		logger: logger,
		table:  internal.GetEnvVar("AWS_DYNAMODB_TABLE_NAME", "zip-requests"),
	}
}

func (repo Repository) Get(ctx context.Context, ref string) (*storage.Entry, error) {
	notFound := storage.NotFoundError{Ref: ref}

	key, _ := attributevalue.Marshal(ref)

	result, err := repo.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &repo.table,
		Key: map[string]types.AttributeValue{
			"Ref": key,
		},
	})
	if err != nil {
		repo.logger.Error(err.Error())
		return nil, notFound
	}

	entry := storage.Entry{}

	err = attributevalue.UnmarshalMap(result.Item, &entry)
	if err != nil {
		repo.logger.Info("Failed to unmarshal Record, ", slog.Any("err", err.Error()))
		return nil, notFound
	}

	if entry.Ref == "" {
		repo.logger.Info("Ref token " + ref + " has expired or does not exist.")
		return nil, notFound
	}

	return &entry, nil
}

func (repo Repository) Delete(ctx context.Context, entry *storage.Entry) error {
	if entry == nil {
		return errors.New("entry cannot be nil")
	}

	key, _ := attributevalue.Marshal(entry.Ref)

	input := &dynamodb.DeleteItemInput{
		TableName: &repo.table,
		Key: map[string]types.AttributeValue{
			"Ref": key,
		},
	}

	_, err := repo.db.DeleteItem(ctx, input)

	return err
}

func (repo Repository) Add(ctx context.Context, entry *storage.Entry) error {
	if entry == nil {
		return errors.New("entry cannot be nil")
	}

	av, err := attributevalue.MarshalMap(entry)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: &repo.table,
		Item:      av,
	}

	_, err = repo.db.PutItem(ctx, input)
	if err != nil {
		return err
	}

	return nil
}
