# opg-file-service

Small microservice built with go to enable users of sirius to download files from s3.

## Local Development
    
### Required Tools

 - Go 1.14.2
 - [GoTestSum](https://github.com/gotestyourself/gotestsum)
 - Docker
 
## Environment Variables


| Variable       | Default                           |  Description   | 
| -------------- | --------------------------------- | -------------- |
| JWT_SECRET     | MyTestSecret                      | Environment variable used to set the key for verifying JWT tokens, this should be overwritten in an environment |
| USER_HASH_SALT | ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0  | Defines what hash to use when hashing user emails, this should match the hash being used on sirius              |

