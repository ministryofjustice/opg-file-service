# opg-file-service

Small microservice built with go to enable users of sirius to download files from s3.

## Local Development
    
### Required Tools

 - Go 1.14.2
 - [GoTestSum](https://github.com/gotestyourself/gotestsum)
 - Docker
 
## Environment Variables


| Variable                  | Default                           |  Description   | 
| ------------------------- | --------------------------------- | -------------- |
| JWT_SECRET                | SeCrEtKeYkNoWnOnLyToMe            | Environment variable used to set the key for verifying JWT tokens, this should be overwritten in an environment |
| USER_HASH_SALT            | ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0  | Defines what hash to use when hashing user emails, this should match the hash being used on sirius              |
| AWS_DYNAMODB_TABLE_NAME   | zip-requests                      | Table name where zip requests are stored                                                                        |
| AWS_S3_ENDPOINT           |                                   | Used for overwriting the S3 endpoint locally e.g. http://localstack:4572                                        |
| AWS_DYNAMODB_ENDPOINT     |                                   | Used for overwriting the DynamoDB endpoint locally e.g. http://localstack:4569                                  |
| AWS_REGION                | eu-west-1                         | Set the AWS region for all operations with the SDK                                                              |
| AWS_ACCESS_KEY_ID         |                                   | Used for authenticating with localstack e.g. set to "localstack"                                                |
| AWS_SECRET_ACCESS_KEY     |                                   | Used for authenticating with localstack e.g. set to "localstack"                                                |
| PATH_PREFIX               |                                   | Path prefix where all requested will be routed                                                                  |
