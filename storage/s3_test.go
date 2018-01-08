package storage

//http://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/s3-example-basic-bucket-operations.html
import (
	"time"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/slotix/dataflowkit/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpload(t *testing.T) {
	uploader := new(mocks.UploaderAPI)
	uploader.On("Upload", mock.AnythingOfType("*s3manager.UploadInput")).Return(&s3manager.UploadOutput{
		Location: "location",
	}, nil)

	err := upload(uploader, "bucket", "key", []byte("value"), 0)
	assert.Nil(t, err, "Expected no error")
}

func TestDownload(t *testing.T) {
	downloader := new(mocks.DownloaderAPI)
	buf := &aws.WriteAtBuffer{}
	//value := []byte("Value")
	//buf.WriteAt(value, int64(len(value)))
	//logger.Info(string(buf.Bytes()))
	//buf := aws.NewWriteAtBuffer([]byte("Value"))

	obj := &s3.GetObjectInput{
		Bucket: aws.String("bucket"),
		Key:    aws.String("key"),
	}
	downloader.On("Download", buf, obj).Return(int64(0), nil)
	_, err := download(downloader, "bucket", "key")
	assert.Nil(t, err, "Expected no error")
}



func TestListBuckets(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("ListBuckets", mock.AnythingOfType("*s3.ListBucketsInput")).Return(&s3.ListBucketsOutput{
		Buckets: []*s3.Bucket{
			&s3.Bucket{Name: aws.String("First Bucket")},
			&s3.Bucket{Name: aws.String("Second Bucket")},
		},
	}, nil)

	b, err := listBuckets(svc)
	assert.Nil(t, err, "Expected no error")
	assert.Len(t, b, 2, "Expect two buckets")
	assert.Equal(t, "First Bucket", b[0], "Expected first bucket")
	assert.Equal(t, "Second Bucket", b[1], "Expected Second Bucket")
}

func TestGetObject(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("GetObject", mock.AnythingOfType("*s3.GetObjectInput")).Return(&s3.GetObjectOutput{}, nil)

	_, err := getObject(svc, "bucket", "key")
	assert.Nil(t, err, "Expected no error")
}

func TestExpiredKey(t *testing.T) {
	expires := "Wed, 22 Nov 2017 15:36:42 GMT"
	lastModified  := time.Now().UTC()
	obj := s3.GetObjectOutput{Expires: &expires, LastModified: &lastModified}
	exp := expiredKey(&obj, int64(3600))
	logger.Info(exp)
}


func TestDelete(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("DeleteObject", mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)

	_, err := delete(svc, "bucket", "key")
	assert.Nil(t, err, "Expected no error")
}

//Actual tests *******
/*
var (
	conn S3Conn
)

func init() {
	viper.Set("SPACES_CONFIG", homeDir()+".spaces/credentials")
	viper.Set("SPACES_ENDPOINT", "https://ams3.digitaloceanspaces.com")
	viper.Set("DFK_BUCKET", "dfk-storage")
	bucket := viper.GetString("DFK_BUCKET")
	config := &aws.Config{
		//Load credentials from specified file
		Credentials: credentials.NewSharedCredentials(viper.GetString("SPACES_CONFIG"), ""),
		//Endpoint is obligatory for DO Spaces
		Endpoint:    aws.String(viper.GetString("SPACES_ENDPOINT")),
		//Region parameter may have any value for Digital Ocean spaces. But it can't be omitted.
		Region:      aws.String("ams333"),
	}
	conn = newS3Conn(config, bucket)
}

func TestListBucketsAWS(t *testing.T) {
	config := &aws.Config{Region: aws.String("us-west-1")}
	sess := session.Must(session.NewSession(config))
	svc := s3.New(sess)
	ifaceClient := s3iface.S3API(svc)
	buckets, err := listBuckets(ifaceClient)
	if err != nil {
		log.Fatalln("Failed to list buckets", err)
	}

	fmt.Println("Buckets:", buckets)
}

func TestListBucketsDO(t *testing.T) {
	sess := session.Must(session.NewSession(conn.config))
	svc := s3.New(sess)
	//result, err := svc.ListBuckets(nil)
	result, err := listBuckets(svc)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Buckets:", result)

	// for _, b := range result.Buckets {
	// 	fmt.Printf("* %s created on %s\n",
	// 		aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	// }

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
	assert := assert.New(t)
	assert.Equal(err, nil)

	//Test download
	downloaded, _ := conn.download("test")
	assert.Equal(buf, downloaded)

	//Test Get Object
	object, _ := conn.getObject("test")
	//	fmt.Println(object.String())
	assert.NotNil(object)

	//Test delete
	_, err = conn.delete("test")
	assert.Nil(err)

	//Test Downloading of item with Invalid Key recently deleted
	downloaded, err = conn.download("test")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			assert.Equal(aerr.Code(), "NoSuchKey")
		}
	}
	//Test Get Object with Invalid Key recently deleted
	_, err = conn.getObject("test")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			assert.Equal(aerr.Code(), "NoSuchKey")
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

*/
//**********
