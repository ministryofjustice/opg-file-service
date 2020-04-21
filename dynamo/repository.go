package dynamo

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

func NewRepository(sess *session.Session, l *log.Logger) *Repository {
	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")

	config := *sess.AwsSession.Config
	config.Endpoint = &endpoint

	dynamo := dynamodb.New(sess.AwsSession, &config)

	return &Repository{
		db:     dynamo,
		logger: l,
	}
}

// TODO: remove ListTables method when done debugging
func (repo Repository) ListTables() {
	// create the input configuration instance
	table := "zip-requests"
	input := &dynamodb.ListTablesInput{
		ExclusiveStartTableName: &table,
	}

	for {
		// Get the list of tables
		repo.logger.Println("endpoint is " + repo.db.Endpoint)
		repo.logger.Println("about to get list of tables...")

		result, err := repo.db.ListTables(input)

		repo.logger.Println(result.TableNames)

		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case dynamodb.ErrCodeInternalServerError:
					repo.logger.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
				default:
					repo.logger.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				repo.logger.Println(err.Error())
			}
			return
		}

		for _, n := range result.TableNames {
			repo.logger.Println(*n)
		}

		// assign the last read tablename as the start for our next call to the ListTables function
		// the maximum number of table names returned in a call is 100 (default), which requires us to make
		// multiple calls to the ListTables function to retrieve all table names
		input.ExclusiveStartTableName = result.LastEvaluatedTableName

		if result.LastEvaluatedTableName == nil {
			break
		}
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
		repo.logger.Println("Failed to unmarshal Record, %v", err)
		return nil, notFound
	}

	if entry.Ref == "" {
		repo.logger.Println("Ref token " + ref + " has expired or does not exist.")
		return nil, notFound
	}

	return &entry, nil
}
