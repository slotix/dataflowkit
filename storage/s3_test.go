package storage

//http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html
import (
	"fmt"
	"log"
	"os/user"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var (
	conn S3Conn
)

func init() {
	viper.Set("SPACES_CONFIG", homeDir()+".spaces/credentials")
	viper.Set("SPACES_ENDPOINT", "https://ams3.digitaloceanspaces.com")
	viper.Set("DFK_BUCKET", "dfk-storage")
	bucket := viper.GetString("DFK_BUCKET")
	config := &aws.Config{
		Credentials: credentials.NewSharedCredentials(viper.GetString("SPACES_CONFIG"), ""), //Load credentials from specified file
		Endpoint:    aws.String(viper.GetString("SPACES_ENDPOINT")),                         //Endpoint is obligatory for DO Spaces
		Region:      aws.String("ams333"),                                                   //Actually for Digital Ocean spaces region parameter may have any value. But it can't be omitted.
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
	sess := session.Must(session.NewSession(conn.config))
	svc := s3.New(sess)
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
	sess := session.Must(session.NewSession(conn.config))
	svc := s3.New(sess)
	resp, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: aws.String(conn.bucket)})

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

func TestUploadDownloadDelete(t *testing.T) {
	//Test upload
	buf := []byte("file content test\nanother line of test here")
	err := conn.upload("test", buf, 0)
	assert.Equal(t, err, nil)

	//Test download
	downloaded, _ := conn.download("test")
	assert.Equal(t, buf, downloaded)

	//Test Get Object
	object, _ := conn.getObject("test")
	//fmt.Println(*object)
	assert.NotNil(t, object)

	//Test delete
	_, err = conn.delete("test")
	assert.Nil(t, err)

	//Test Downloading of item with Invalid Key recently deleted
	downloaded, err = conn.download("test")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			assert.Equal(t, aerr.Code(), "NoSuchKey")
		}
	}
	//Test Get Object with Invalid Key recently deleted
	_, err = conn.getObject("test")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			assert.Equal(t, aerr.Code(), "NoSuchKey")
		}
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
