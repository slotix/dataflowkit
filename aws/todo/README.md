# Serverless AWS Golang Event 
Serverless AWS APIGateway events example using: 

- [AWS Lambda Go Shim](https://github.com/eawsy/aws-lambda-go-shim)
- [API Gateway Proxy Event](https://github.com/eawsy/aws-lambda-go-event/tree/master/service/lambda/runtime/event/apigatewayproxyevt)

Each CRUD operation is it's own Lambda Function. This is convenient to hook into other 
Event Source triggers such as `Kinesis` or `SNS`.

## Usage
Setup and deploy a new project called `your-app`:

```bash
cd $GOPATH/src/your-path/
serverless install -u https://github.com/yunspace/serverless-golang/tree/master/examples/aws-golang-event -n your-app
```

```bash
cd 
make DOTENV=.env.example dotenv

```
* fill in and correct any of the variables in .env
* replace `WORKDIR` in .env with `/go/src/your-path/your-app`

```bash
make test build deploy
```

use curl, PostMan or any REST client and do a `POST` on the provided gateway endpoint:

```json
{
  "id": "c576e9bc-548e-457d-b662-254ade7fc695",
  "message": "hello world"
}
```

should get the same payload back with status `200`

```bash
make remove
```
