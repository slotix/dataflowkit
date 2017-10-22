package fetch

import (
	"bytes"

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
	sqsSvc := sqs.New(sess)

	//fetch results if there is nothing in a cache
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	if sResponse, ok := resp.(*splash.Response); ok {
		//if sResponse.Cacheable {
		content, err := sResponse.GetContent()
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		n, _ := buf.ReadFrom(content)
		logger.Printf("URL: %s Size : %.2f kb\n",mw.getURL(req), float64(n)/float64(1024))
		strContent := buf.String() 
		

		params := &sqs.SendMessageInput{
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"URL": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(mw.getURL(req)),
				},
			},
			MessageBody:  aws.String(strContent),                                          // Required
			QueueUrl:     aws.String(viper.GetString("SQS_QUEUE_FETCH_URL_OUT")), // Required
			DelaySeconds: aws.Int64(0),                                          // (optional)  0 ~ 900s (15 minutes)
		}
		_, err = sqsSvc.SendMessage(params)
		if err != nil {
			return nil, err
		}

		//	}
		output = sResponse
	}
	//output, err = sResponse.GetContent()
	return
}
