# Serverless AWS Golang Net
Serverless AWS Go Net example using:
 
- [AWS Lambda Golang Net](https://github.com/eawsy/aws-lambda-go-net). 
- [Gorilla MUX](http://www.gorillatoolkit.org/pkg/mux) 

All CRUD operations are within a single Lambda Function behind `gorilla/mux`. This example is
very handy for porting over existing `golang/net` projects. However it is not readily compatible 
with other Event Source triggers such as `Kinesis` or `SNS`.

## Usage
Setup and deploy a new project called `your-app`:

```bash
cd $GOPATH/src/your-path/
serverless install -u https://github.com/yunspace/serverless-golang/tree/master/examples/aws-golang-event -n your-app
```

```bash
cd you-app
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
