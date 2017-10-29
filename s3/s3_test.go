package s3

//http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html
import (
	"bytes"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	svc        *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	bucket string
)

func init() {
	bucket = "fetch-bucket"
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc = s3.New(sess)
	uploader = s3manager.NewUploader(sess)
	downloader = s3manager.NewDownloader(sess)
	
}

func TestListBuckets(t *testing.T) {
	result, err := svc.ListBuckets(nil)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

}

func TestListBucketItems(t *testing.T) {
	
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})

	if err != nil {
		fmt.Println(err)
	}

	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

func TestUpload(t *testing.T) {
	
	buf := []byte("file content test\nanother line of test here")
	r := bytes.NewReader(buf)
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("urlll"),
		Body:   r,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Type: %T\n",err)
	}

}

func TestDownload(t *testing.T) {
	
//	var buf []byte
	buff := &aws.WriteAtBuffer{}

//	numBytes, err := downloader.Download(aws.NewWriteAtBuffer(buf),
	numBytes, err := downloader.Download(buff,
	&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String("urlll"),
		})

	if err != nil {
		fmt.Println(err.(s3.RequestFailure))
		fmt.Println(err.(s3.RequestFailure).Code())
		fmt.Printf("Type: %T\n",err)
	}

	fmt.Printf("Content: %s\n", string(buff.Bytes()))
	fmt.Println("Downloaded", numBytes, "bytes")

}

func TestDeleteItem(t *testing.T) {
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String("urlll")})
	
	if err != nil {
		fmt.Println(err)
	}
	
		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String("urlll"),
		})
		if err != nil {
			fmt.Println(err)
		}
	
}