package storage

import (
	"bytes"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Conn represents a AWS S3 Connection structure
type S3Conn struct {
	//bucket name
	bucket     string
	svc        *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}


/* func newS3Conn(bucket string) S3Conn {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	return S3Conn{bucket, svc, uploader, downloader}
} */


// newS3Conn initializes new AWS S3 / Digital Ocean Spaces Connection with specified bucket
//load credentials from shared file
//credentials have the following format:
//[default]
//aws_access_key_id = some_access_key_id
//aws_secret_access_key = some_secret_access_key
//--------
//Spaces access keys are generated in DO Control panel at
//https://cloud.digitalocean.com/settings/api/tokens?i=2c1aad
func newS3Conn(config *aws.Config, bucket string) S3Conn {
	sess := session.New(config)
	svc := s3.New(sess)

	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	return S3Conn{bucket, svc, uploader, downloader}
}

//download returns a value of specified key from AWS S3
func (s S3Conn) download(key string) (value []byte, err error) {
	buf := &aws.WriteAtBuffer{}
	_, err = s.downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(s.bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//upload sends key/value pair to AWS S3 Storage
func (s S3Conn) upload(key string, value []byte, expTime int64) error {
	r := bytes.NewReader(value)
	t := time.Unix(expTime, 0)
	_, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(key),
		Body:    r,
		Expires: &t,
	})
	if err != nil {
		return err
	}
	return nil
}

//getObject returns an object from AWS S3. This may be used to get meta information about an object.
func (s S3Conn) getObject(key string) (object *s3.GetObjectOutput, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}
	object, err = s.svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}
	return
}
