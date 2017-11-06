package storage

import (
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Conn struct {
	bucket     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
}

func newS3Conn(bucket string) S3Conn {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	return S3Conn{bucket, uploader, downloader}
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
