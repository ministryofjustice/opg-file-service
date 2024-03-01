package dynamo

import (
	"bytes"
	"errors"
	"log/slog"
	"opg-file-service/session"
	"opg-file-service/storage"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/stretchr/testify/assert"
)

func str(s string) *string {
	return &s
}

func TestNewRepository(t *testing.T) {
	tests := []struct {
		scenario     string
		endpoint     *string
		table        *string
		wantEndpoint string
		wantTable    string
	}{
		{
			"New Repository with default config",
			nil,
			nil,
			"https://dynamodb.eu-west-1.amazonaws.com",
			"zip-requests",
		},
		{
			"New Repository with config overrides",
			str("http://localhost"),
			str("test"),
			"http://localhost",
			"test",
		},
	}

	for _, test := range tests {
		sess, _ := session.NewSession()
		var buf bytes.Buffer
		l := slog.New(slog.NewJSONHandler(&buf, nil))

		os.Unsetenv("AWS_DYNAMODB_ENDPOINT")
		os.Unsetenv("AWS_DYNAMODB_TABLE_NAME")
		if test.endpoint != nil {
			os.Setenv("AWS_DYNAMODB_ENDPOINT", *test.endpoint)
		}
		if test.table != nil {
			os.Setenv("AWS_DYNAMODB_TABLE_NAME", *test.table)
		}

		repo := NewRepository(*sess, l)

		assert.Equal(t, test.wantTable, repo.table, test.scenario)
		assert.Equal(t, l, repo.logger, test.scenario)
		assert.IsType(t, new(dynamodb.DynamoDB), repo.db, test.scenario)
		db := repo.db.(*dynamodb.DynamoDB)
		assert.Equal(t, test.wantEndpoint, db.Endpoint, test.scenario)
	}
}

func newValidGetItemOutput(ref string) *dynamodb.GetItemOutput {
	item := map[string]*dynamodb.AttributeValue{
		"Files": {
			L: []*dynamodb.AttributeValue{
				{
					M: map[string]*dynamodb.AttributeValue{
						"S3Path": {
							S: aws.String("s3://files/file.test"),
						},
						"FileName": {
							S: aws.String("file.test"),
						},
						"Folder": {
							S: aws.String("folder"),
						},
					},
				},
			},
		},
		"Hash": {
			S: aws.String("testHash"),
		},
		"Ref": {
			S: &ref,
		},
		"Ttl": {
			N: aws.String("0"),
		},
	}

	return &dynamodb.GetItemOutput{
		Item: item,
	}
}

func TestRepository_Get(t *testing.T) {
	tests := []struct {
		scenario  string
		ref       string
		dbOut     *dynamodb.GetItemOutput
		dbErr     error
		wantOut   *storage.Entry
		wantErr   error
		wantInLog string
	}{
		{
			"Successfully retrieve and un-marshall data from DB.",
			"test",
			newValidGetItemOutput("test"),
			nil,
			&storage.Entry{
				Ref:  "test",
				Hash: "testHash",
				Ttl:  0,
				Files: []storage.File{
					{
						S3path:   "s3://files/file.test",
						FileName: "file.test",
						Folder:   "folder",
					},
				},
			},
			nil,
			"",
		},
		{
			"Error from DB client when retrieving item.",
			"test",
			nil,
			errors.New("some DB error"),
			nil,
			storage.NotFoundError{Ref: "test"},
			"some DB error",
		},
		{
			"Entry not found.",
			"test",
			new(dynamodb.GetItemOutput),
			nil,
			nil,
			storage.NotFoundError{Ref: "test"},
			"Ref token test has expired or does not exist",
		},
	}

	for _, test := range tests {
		var buf bytes.Buffer
		l := slog.New(slog.NewJSONHandler(&buf, nil))
		mdb := MockDynamoDB{}

		repo := Repository{
			db:     &mdb,
			logger: l,
			table:  "table",
		}

		input := dynamodb.GetItemInput{
			TableName: &repo.table,
			Key: map[string]*dynamodb.AttributeValue{
				"Ref": {
					S: &test.ref,
				},
			},
		}

		mdb.On("GetItem", &input).Return(test.dbOut, test.dbErr).Once()

		entry, err := repo.Get(test.ref)

		assert.Equal(t, test.wantErr, err, test.scenario)
		assert.Equal(t, test.wantOut, entry, test.scenario)
		assert.Contains(t, buf.String(), test.wantInLog, test.scenario)
	}
}

func TestRepository_Delete(t *testing.T) {
	tests := []struct {
		scenario string
		entry    *storage.Entry
		dbErr    error
		wantErr  error
	}{
		{
			"Successful delete",
			&storage.Entry{Ref: "test"},
			nil,
			nil,
		},
		{
			"Error from DB client",
			&storage.Entry{Ref: "test"},
			errors.New("some DB error"),
			errors.New("some DB error"),
		},
		{
			"Nil entry parameter",
			nil,
			nil,
			errors.New("entry cannot be nil"),
		},
	}

	for _, test := range tests {
		mdb := MockDynamoDB{}

		var buf bytes.Buffer
		repo := Repository{
			db:     &mdb,
			logger: slog.New(slog.NewJSONHandler(&buf, nil)),
			table:  "table",
		}

		ref := ""
		if test.entry != nil {
			ref = test.entry.Ref
		}

		input := dynamodb.DeleteItemInput{
			TableName: &repo.table,
			Key: map[string]*dynamodb.AttributeValue{
				"Ref": {
					S: &ref,
				},
			},
		}

		mdb.On("DeleteItem", &input).Return(new(dynamodb.DeleteItemOutput), test.dbErr).Once()

		err := repo.Delete(test.entry)

		assert.Equal(t, test.wantErr, err, test.scenario)
	}
}

func TestRepository_Add(t *testing.T) {
	tests := []struct {
		scenario string
		entry    *storage.Entry
		dbErr    error
		wantErr  error
	}{
		{
			"Entry added successfully",
			&storage.Entry{},
			nil,
			nil,
		},
		{
			"Error from DB client",
			&storage.Entry{},
			errors.New("some DB error"),
			errors.New("some DB error"),
		},
		{
			"Nil entry parameter",
			nil,
			nil,
			errors.New("entry cannot be nil"),
		},
	}

	for _, test := range tests {
		mdb := MockDynamoDB{}

		var buf bytes.Buffer
		repo := Repository{
			db:     &mdb,
			logger: slog.New(slog.NewJSONHandler(&buf, nil)),
			table:  "table",
		}

		av, _ := dynamodbattribute.MarshalMap(test.entry)
		input := dynamodb.PutItemInput{
			TableName: &repo.table,
			Item:      av,
		}

		mdb.On("PutItem", &input).Return(new(dynamodb.PutItemOutput), test.dbErr).Once()

		err := repo.Add(test.entry)

		assert.Equal(t, test.wantErr, err, test.scenario)
	}
}
