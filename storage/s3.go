package storage

import (
	"bytes"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
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

//download returns a value of specified key from AWS S3 bucket.
// s3manageriface.DownloaderAPI interface helps to mock real S3 connection.
func download(d s3manageriface.DownloaderAPI, bucket, key string) ([]byte, error) {
	//downloader := s3manager.NewDownloaderWithClient(svc)
	buf := &aws.WriteAtBuffer{}
	object := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	_, err := d.Download(buf, object)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//download returns a value of specified key from AWS S3
func (c S3Conn) download(key string) (value []byte, err error) {
	sess := session.Must(session.NewSession(c.config))
	//svc := s3.New(sess)
	downloader := s3manager.NewDownloader(sess)
	value, err = download(downloader, c.bucket, key)

	return
}

//upload uploads key/value pair to AWS S3 Storage and sets up Expires parameter
// s3manageriface.UploaderAPI helps to mock real S3 connection.
func upload(u s3manageriface.UploaderAPI, bucket, key string, value []byte, expTime int64) error {
	//uploader := s3manager.NewUploaderWithClient(svc)
	r := bytes.NewReader(value)
	t := time.Unix(expTime, 0)
	_, err := u.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		Body:    r,
		Expires: &t,
	})
	if err != nil {
		return err
	}
	return nil
}

//upload uploads key/value pair to AWS S3 Storage and sets up Expires parameter
func (c S3Conn) upload(key string, value []byte, expTime int64) error {
	sess := session.Must(session.NewSession(c.config))
	//svc := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	err := upload(uploader, c.bucket, key, value, expTime)
	return err
}

//getObject returns an object from AWS S3. This may be used to get meta information about an object.
func (c S3Conn) getObject(key string) (object *s3.GetObjectOutput, err error) {
	sess := session.Must(session.NewSession(c.config))
	svc := s3.New(sess)
	object, err = getObject(svc, c.bucket, key)
	return
}

//getObject returns an object from AWS S3. This may be used to get meta information about an object.
func getObject(svc s3iface.S3API, bucket, key string) (object *s3.GetObjectOutput, err error) {
	//sess := session.Must(session.NewSession(c.config))
	//svc := s3.New(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	object, err = svc.GetObject(input)
	return
}

func listBuckets(svc s3iface.S3API) ([]string, error) {
	resp, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]string, 0, len(resp.Buckets))
	for _, b := range resp.Buckets {
		buckets = append(buckets, *b.Name)
	}

	return buckets, nil
}

func delete(svc s3iface.S3API, bucket, key string) (result *s3.DeleteObjectOutput, err error) {
	//sess := session.Must(session.NewSession(c.config))
	//svc := s3.New(sess)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	result, err = svc.DeleteObject(input)
	return
}

func (c S3Conn) delete(key string) (result *s3.DeleteObjectOutput, err error) {
	sess := session.Must(session.NewSession(c.config))
	svc := s3.New(sess)
	result, err = delete(svc, c.bucket, key)
	return
}
