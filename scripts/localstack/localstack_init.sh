#! /usr/bin/env bash

# Create a DynamoDB table
awslocal dynamodb create-table \
  --table-name zip-requests \
  --attribute-definitions AttributeName=ref,AttributeType=S \
  --key-schema AttributeName=ref,KeyType=HASH \
  --provisioned-throughput ReadCapacityUnits=1000,WriteCapacityUnits=1000

# Set automatic expiry of records based on the ttl attribute
awslocal dynamodb update-time-to-live \
  --table-name zip-requests \
  --time-to-live-specification "Enabled=true, AttributeName=ttl"

# Create an S3 bucket
awslocal s3 mb s3://files

# Put some files in S3 for testing
touch dummy.txt
echo "File content" > dummy.txt
for i in {1..3}
do
  awslocal s3 cp dummy.txt "s3://files/file$i.txt"
done

# Add some records in DynamoDB for testing. Hash is for Test.McTestFace@mail.com
item='{"ref": {"S":"REF_GOES_HERE"},"ttl": {"N":"9999999999"},"hash": {"S":"d1a046e6300ea9a75cc4f9eda85e8442c3e9913b8eeb4ed0895896571e479a99"},"files": {"L": [{"M": {"S3path": {"S":"s3://files/file1.txt"},"FileName": {"S":"file1.txt"}}},{"M": {"S3path": {"S":"s3://files/file2.txt"},"FileName": {"S":"file2.txt"}}},{"M": {"S3path": {"S":"s3://files/file3.txt"},"FileName": {"S":"file3.txt"},"Folder": {"S":"folder"}}}]}}'
for i in {1..10}
do
  awslocal dynamodb put-item --table-name zip-requests --item "${item/REF_GOES_HERE/ref$i}"
done
