#! /usr/bin/env bash

# Create a DynamoDB table
awslocal dynamodb create-table \
  --table-name zip-requests \
  --attribute-definitions AttributeName=Ref,AttributeType=S \
  --key-schema AttributeName=Ref,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

# Set automatic expiry of records based on the ttl attribute
awslocal dynamodb update-time-to-live \
  --table-name zip-requests \
  --time-to-live-specification "Enabled=true, AttributeName=Ttl"
