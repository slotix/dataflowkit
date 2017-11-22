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

type S3Conn struct {
	bucket     string
	svc        *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func newS3Conn(bucket string) S3Conn {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	return S3Conn{bucket, svc, uploader, downloader}
}

func (s S3Conn) Download(key string) (value []byte, err error) {
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

func (s S3Conn) GetObject(key string) (object *s3.GetObjectOutput, err error) {
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
	//expires, err = time.Parse("Mon, 2 Jan 2006 15:04:05 MST", *result.Expires)
	//if err != nil {
	//	return time.Time{}, err
	//}
	//return expires.UTC(), nil

}



func (s S3Conn) Upload(key string, value []byte, expTime int64) error {
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
