package fetch

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

type sqsMiddleware struct {
	Service
}

// implement function to return ServiceMiddleware
func SQSMiddleware() ServiceMiddleware {
	return func(next Service) Service {
		return sqsMiddleware{next}
	}
}

//var redisCon cache.RedisConn

func (mw sqsMiddleware) Fetch(req interface{}) (output interface{}, err error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(viper.GetString("SQS_AWS_REGION")),
		//	Credentials: credentials.NewSharedCredentials("/Users/dm/.aws/credentials", "default"),
		MaxRetries: aws.Int(viper.GetInt("SQS_MAX_RETRIES")),
	})
	if err != nil {
		return nil, err
	}
	logger.Println(viper.GetInt("SQS_MAX_RETRIES"))
	sqsSvc := sqs.New(sess)

	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	if sResponse, ok := resp.(*splash.Response); ok {
		if sResponse.Cacheable {

			params := &sqs.SendMessageInput{
				MessageBody:  aws.String(mw.getURL(req)),                   // Required
				QueueUrl:     aws.String(viper.GetString("SQS_QUEUE_URL")), // Required
				DelaySeconds: aws.Int64(0),                                 // (optional)  0 ~ 900s (15 minutes)
			}
			_, err := sqsSvc.SendMessage(params)
			if err != nil {
				return nil, err
			}

		}
		output = sResponse
	}
	//output, err = sResponse.GetContent()
	return
}
