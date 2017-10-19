package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	QueueUrl    = "https://sqs.us-east-1.amazonaws.com/060679207441/fetch"
	AWSRegion   = "us-east-1"
	CredPath    = "/Users/dm/.aws/credentials"
	CredProfile = "default"
	MaxRetries  = 5
)

func NewSvc(Region, CredPath, CredProfile string, MaxRetries int) *sqs.SQS {
	sess := session.New(&aws.Config{
		Region:      aws.String(Region),
		Credentials: credentials.NewSharedCredentials(CredPath, CredProfile),
		MaxRetries:  aws.Int(MaxRetries),
	})

	svc := sqs.New(sess)
	return svc
}
func ListQueuses(svc *sqs.SQS) ([]*string, error){
	result, err := svc.ListQueues(nil)
	if err != nil {
		return nil, err
	}
	return result.QueueUrls, nil
	
}

// SendMessage send message to QueueUrl
//returned value is 
//{
//	MD5OfMessageBody: "d1d4180b7e411c4be86b00fb2ee103eb",
//	MessageId: "9e398e4e-f157-45bf-a9b8-f6cca48fd1f6"
// }
//or error if any 
func SendMessage(svc *sqs.SQS, queueUrl, message string, delay int64) (*sqs.SendMessageOutput, error){
	params := &sqs.SendMessageInput{
		MessageBody:  aws.String(message),  // Required
		QueueUrl:     aws.String(queueUrl), // Required
		DelaySeconds: aws.Int64(delay),         // (optional)  0 ~ 900s (15 minutes)
	}
	output, err := svc.SendMessage(params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Receive message
func ReceiveMessage(svc *sqs.SQS, queueUrl string, maxNumberOfMessages, visibilityTimeout, waitTimeSeconds int64) (*sqs.ReceiveMessageOutput, error){
	params := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(maxNumberOfMessages),
		VisibilityTimeout:   aws.Int64(visibilityTimeout),
		WaitTimeSeconds:     aws.Int64(waitTimeSeconds),
	}
	output, err := svc.ReceiveMessage(params)
	if err != nil {
		return nil, err
	}
	return output, nil
}



func DeleteMessage(svc *sqs.SQS, receive_resp *sqs.ReceiveMessageOutput) {
	// Delete message
	for _, message := range receive_resp.Messages {
		delete_params := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(QueueUrl),  // Required
			ReceiptHandle: message.ReceiptHandle, // Required
		}
		_, err := svc.DeleteMessage(delete_params) // No response returned when successed.
		if err != nil {
			logger.Println(err)
		}
		logger.Printf("[Delete message] \nMessage ID: %s has beed deleted.\n\n", *message.MessageId)
	}
}

//purgeQueue deletes all messages from a queue.
func purgeQueue(svc *sqs.SQS, input *sqs.PurgeQueueInput) error {
	_, err := svc.PurgeQueue(input)
	if err != nil {
		return err
	}
	return nil
}
