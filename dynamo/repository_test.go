package dynamo

import (
	"bytes"
	"errors"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"opg-file-service/storage"
	"testing"
)

func newValidGetItemOutput(ref string) *dynamodb.GetItemOutput {
	entry := storage.Entry{
		Ref:  ref,
		Hash: "testHash",
		Ttl:  0,
		Files: []storage.File{
			{
				S3path:   "s3://files/file.test",
				FileName: "file.test",
				Folder:   "folder",
			},
		},
	}

	item, err := attributevalue.MarshalMap(entry)
	if err != nil {
		panic(err)
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

		key, _ := attributevalue.Marshal(test.ref)

		input := dynamodb.GetItemInput{
			TableName: &repo.table,
			Key: map[string]types.AttributeValue{
				"Ref": key,
			},
		}

		mdb.On("GetItem", &input).Return(test.dbOut, test.dbErr).Once()

		entry, err := repo.Get(t.Context(), test.ref)

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

		key, _ := attributevalue.Marshal(ref)

		input := dynamodb.DeleteItemInput{
			TableName: &repo.table,
			Key: map[string]types.AttributeValue{
				"Ref": key,
			},
		}

		mdb.On("DeleteItem", &input).Return(new(dynamodb.DeleteItemOutput), test.dbErr).Once()

		err := repo.Delete(t.Context(), test.entry)

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

		av, _ := attributevalue.MarshalMap(test.entry)
		input := dynamodb.PutItemInput{
			TableName: &repo.table,
			Item:      av,
		}

		mdb.On("PutItem", &input).Return(new(dynamodb.PutItemOutput), test.dbErr).Once()

		err := repo.Add(t.Context(), test.entry)

		assert.Equal(t, test.wantErr, err, test.scenario)
	}
}
