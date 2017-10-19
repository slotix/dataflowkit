package sqs

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
)

var svc *sqs.SQS

func init() {
	svc = NewSvc(
		AWSRegion,
		CredPath,
		CredProfile,
		MaxRetries,
	)
}
func TestListQueuses(t *testing.T) {

	urls, err := ListQueuses(svc)
	if err != nil {
		logger.Println(err)
	}
	// As these are pointers, printing them out directly would not be useful.
	for i, url := range urls {
		// Avoid dereferencing a nil pointer.
		if url == nil {
			continue
		}
		logger.Printf("%d: %s\n", i, *url)
	}

}

func TestSendMessage(t *testing.T) {
	message := "Test223 Message"
	output, err := SendMessage(svc,
		QueueUrl,
		message,
		0)
	if err != nil {
		logger.Println(err)
	}
	logger.Printf("[Send message] \n%v \n\n", output)
}

func TestReceiveMessage(t *testing.T) {
	output, err := ReceiveMessage(svc,
		QueueUrl,
		6,  //MaxNumberOfMessages
		15, //VisibilityTimeout
		10) //WaitTimeSeconds
	if err != nil {
		logger.Println(err)
	}
	logger.Printf("[Receive message] \n%v \n\n", output)
}

func TestPurgeQueue(t *testing.T) {
	qURL := QueueUrl
	input := sqs.PurgeQueueInput{
		QueueUrl: &qURL,
	}

	err := purgeQueue(svc, &input)
	if err != nil {
		logger.Println(err)
	}
}
