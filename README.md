# rally-github-service

## Description
The purpose of this service is to replace the rally service that is currently integrated with GitHub and scheduled for deprecation.

The service is a rewrite/port of the [GitHub Rally service](https://github.com/github/github-services/blob/master/lib/services/rally.rb) 

## Usage
The service currently looks for a small json configuration file in the working directory on start-up
```json
{ 
    "rally-url": "<add your rally url here>",
    "api-key": "<add your rally api key here>",
    "workspace": "<add your workspace here>",
    "signature_required": false,
    "secret_token": "add your secret GitHub token"
}
```
**rally-url:** The url to your rally server.  
**api-key:** Your rally API key  
**workspace:** Your rally workspace  
**signature_required:** Set true if payloads are required to be signed by a secret token  
**secret_token:** Token used to generate the HMAC hash when signing the payload.

**Note:** If using secrets on GitHub to sign payloads you will need to generate the secret. Instructions are on Github [here](https://developer.github.com/webhooks/securing/#setting-your-secret-token).  

### Setting the hook
1. Navigate to your organizarion or repository.
2. Select settings -> hooks, you will need to have admin permissions.
3. Select "Add webhook".
4. In the Payload URL field enter the url to your webhook deployment.
5. Enter a secret if desired.
6. Click "Add webhook".

### Commit Message Format

The hook will parse out Rally ID's from the commit message in the format of upper case D|DE|DS|TA|TC|S|US followed by a number of digits. The hook will also parse verbs in the form of STARTS|BEGINS and COMPLETES|FINISHES if they precede the rally identifier.

```sh
git commit -m "STARTS US12345 - this is a commit message"
```
The above commit message will attach a changeset and update the status of user story `US12345` to `In Progress`.

## Development
### Prerequisites
The project has been tested with Go 1.12.3

### Test
1. Clone or download the source from GitHub.
2. The project uses build tags to separate unit and integration tests and can be run as described below.
```sh
go test -v ./... -tags unit
```
or
```sh
go test -v ./... -tags integration
```

### Build
Building is done with a standard Go build.
```sh
go build -o rally-github-service server/main.go
```

## License

Refer to [LICENSE](LICENSE.md)
