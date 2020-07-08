package dynamo

import (
	"bytes"
	"errors"
	"log"
	"opg-file-service/session"
	"opg-file-service/storage"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		scenario     string
		endpoint     string
		table        string
		wantEndpoint string
		wantTable    string
	}{
		{
			scenario:     "default_config",
			table:        "zip-requests",
			wantEndpoint: "https://dynamodb.eu-west-1.amazonaws.com",
			wantTable:    "zip-requests",
		},
		{
			scenario:     "overrides",
			endpoint:     "http://localhost",
			table:        "test",
			wantEndpoint: "http://localhost",
			wantTable:    "test",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			sess, _ := session.NewSession("eu-west-1", "")
			var buf bytes.Buffer
			l := log.New(&buf, "test", log.LstdFlags)

			repo := NewRepository(*sess, l, test.endpoint, test.table)

			assert.Equal(t, test.wantTable, repo.table)
			assert.Equal(t, l, repo.logger)
			assert.IsType(t, new(dynamodb.DynamoDB), repo.db)
			db := repo.db.(*dynamodb.DynamoDB)
			assert.Equal(t, test.wantEndpoint, db.Endpoint)
		})
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
			scenario: "success",
			ref:      "test",
			dbOut:    newValidGetItemOutput("test"),
			wantOut: &storage.Entry{
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
		},
		{
			scenario:  "error_from_db",
			ref:       "test",
			dbErr:     errors.New("some DB error"),
			wantErr:   storage.NotFoundError{Ref: "test"},
			wantInLog: "some DB error",
		},
		{
			scenario:  "entry_not_found",
			ref:       "test",
			dbOut:     new(dynamodb.GetItemOutput),
			wantErr:   storage.NotFoundError{Ref: "test"},
			wantInLog: "Ref token test has expired or does not exist",
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			var buf bytes.Buffer
			l := log.New(&buf, "", log.LstdFlags)

			mdb := &MockDynamoDB{}

			repo := &Repository{
				db:     mdb,
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

			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantOut, entry)
			assert.Contains(t, buf.String(), test.wantInLog)
		})
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

		repo := Repository{
			db:     &mdb,
			logger: &log.Logger{},
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
			"successfully_add_entry",
			&storage.Entry{},
			nil,
			nil,
		},
		{
			"db_client_error",
			&storage.Entry{},
			errors.New("some DB error"),
			errors.New("some DB error"),
		},
		{
			"nil_entry",
			nil,
			nil,
			errors.New("entry cannot be nil"),
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			db := &MockDynamoDB{}

			repo := Repository{
				db:     db,
				logger: &log.Logger{},
				table:  "table",
			}

			av, _ := dynamodbattribute.MarshalMap(test.entry)
			input := dynamodb.PutItemInput{
				TableName: &repo.table,
				Item:      av,
			}

			db.On("PutItem", &input).
				Return(new(dynamodb.PutItemOutput), test.dbErr).
				Once()

			err := repo.Add(test.entry)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

type MockDynamoDB struct {
	mock.Mock
}

func (m *MockDynamoDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDB) DeleteItem(input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *MockDynamoDB) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func str(s string) *string {
	return &s
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
