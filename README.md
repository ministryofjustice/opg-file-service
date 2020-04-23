# opg-s3-zipper-service
Golang microservice to pull documents from s3 and zip them: Managed by opg-org-infra &amp; Terraform

## Local Development

## Environment Variables


| Variable       | Default                           |  Description   | 
| -------------- | --------------------------------- | -------------- |
| JWT_SECRET     | MyTestSecret                      | Environment variable used to set the key for verifying JWT tokens, this should be overwritten in an environment |
| USER_HASH_SALT | ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0  | Defines what hash to use when hashing user emails, this should match the hash being used on sirius              |


