package storage

import (
	"bytes"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//S3Conn represents a AWS S3 Connection structure
type S3Conn struct {
	config *aws.Config
	bucket string
}

// newS3Conn initializes new AWS S3 / Digital Ocean Spaces
//load credentials from shared file
//credentials have the following format:
//[default]
//aws_access_key_id = some_access_key_id
//aws_secret_access_key = some_secret_access_key
//--------
//Spaces access keys are generated in DO Control panel at
//https://cloud.digitalocean.com/settings/api/tokens?i=2c1aad
func newS3Conn(config *aws.Config, bucket string) S3Conn {
	return S3Conn{config, bucket}
}

//download returns a value of specified key from AWS S3
func (c S3Conn) download(key string) (value []byte, err error) {
	sess := session.Must(session.NewSession(c.config))
	downloader := s3manager.NewDownloader(sess)
	buf := &aws.WriteAtBuffer{}
	_, err = downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(c.bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//upload sends key/value pair to AWS S3 Storage
func (c S3Conn) upload(key string, value []byte, expTime int64) error {
	sess := session.Must(session.NewSession(c.config))
	uploader := s3manager.NewUploader(sess)
	r := bytes.NewReader(value)
	t := time.Unix(expTime, 0)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(c.bucket),
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
func (c S3Conn) getObject(key string) (object *s3.GetObjectOutput, err error) {
	sess := session.Must(session.NewSession(c.config))
	svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}
	object, err = svc.GetObject(input)
	return
}

func (c S3Conn) delete(key string) (result *s3.DeleteObjectOutput, err error) {
	sess := session.Must(session.NewSession(c.config))
	svc := s3.New(sess)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	result, err = svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				return result, errors.New(aerr.Error())

			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return result, errors.New(aerr.Error())
		}
		return
	}
	return
}
