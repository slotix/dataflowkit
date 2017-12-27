package storage

//http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html
import (
	"bytes"
	"fmt"
	"log"
	"os/user"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
)

var (
	//svc        *s3.S3
	//uploader   *s3manager.Uploader
	//downloader *s3manager.Downloader
	bucket string
	conn   S3Conn
)

func init() {
	viper.Set("SPACES_CONFIG", homeDir()+".spaces/credentials")
	viper.Set("SPACES_ENDPOINT", "https://ams3.digitaloceanspaces.com")
	bucket = viper.GetString("DFK_BUCKET")
	config := &aws.Config{
		Credentials: credentials.NewSharedCredentials(viper.GetString("SPACES_CONFIG"), ""), //Load credentials from specified file
		Endpoint:    aws.String(viper.GetString("SPACES_ENDPOINT")),                         //Endpoint is obligatory for DO Spaces
		Region:      aws.String("ams333"),                                                   //Actually for Digital Ocean spaces region parameter may have any value. But it can't be omited.
	}

	conn = newS3Conn(config, bucket)
	/* bucket = "fetch-bucket"
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create S3 service client
	svc = s3.New(sess)
	uploader = s3manager.NewUploader(sess)
	downloader = s3manager.NewDownloader(sess) */

}

func TestListBuckets(t *testing.T) {
	result, err := conn.svc.ListBuckets(nil)

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

	resp, err := conn.svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(bucket)})

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
	_, err := conn.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("urlll"),
		Body:   r,
	})
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Type: %T\n", err)
	}

}

func TestDownload(t *testing.T) {

	//	var buf []byte
	buff := &aws.WriteAtBuffer{}

	//	numBytes, err := downloader.Download(aws.NewWriteAtBuffer(buf),
	numBytes, err := conn.downloader.Download(buff,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String("http://dbconvert.com"),
		})

	if err != nil {
		fmt.Println(err.(s3.RequestFailure))
		fmt.Println(err.(s3.RequestFailure).Code())
		fmt.Printf("Type: %T\n", err)
	}

	fmt.Printf("Content: %s\n", string(buff.Bytes()))
	fmt.Println("Downloaded", numBytes, "bytes")

}

func TestGetObject(t *testing.T) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("test"),
	}

	result, err := conn.svc.GetObject(input)
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

	fmt.Println(*result)
	//expires, err := time.Parse("Mon, 2 Jan 2006 15:04:05 MST", *result.Expires)
	//fmt.Println(expires.UTC())
}

func TestDeleteItem(t *testing.T) {
	_, err := conn.svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String("urlll")})

	if err != nil {
		fmt.Println(err)
	}

	err = conn.svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("urlll"),
	})
	if err != nil {
		fmt.Println(err)
	}

}

//homeDir returns user's $HOME directory
func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	return usr.HomeDir + "/"
}
