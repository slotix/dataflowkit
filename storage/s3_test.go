package storage

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/slotix/dataflowkit/storage/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_upload(t *testing.T) {
	uploader := new(mocks.UploaderAPI)
	uploader.On("Upload", mock.AnythingOfType("*s3manager.UploadInput")).Return(&s3manager.UploadOutput{
		Location: "location",
	}, nil)

	err := upload(uploader, "bucket", "key", []byte("value"), 0)
	assert.Nil(t, err, "Expected no error")
}

func Test_download(t *testing.T) {
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

func Test_listBuckets(t *testing.T) {
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

func Test_getObject(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("GetObject", mock.AnythingOfType("*s3.GetObjectInput")).Return(&s3.GetObjectOutput{}, nil)

	_, err := getObject(svc, "bucket", "key")
	assert.Nil(t, err, "Expected no error")
}

func Test_expiredKey(t *testing.T) {
	modifiedRightNow := time.Now().UTC()
	expired := expiredKey(&s3.GetObjectOutput{
		LastModified: &modifiedRightNow}, 
		int64(3600))
	assert.Equal(t, expired, false, "Expected false for an item modified right now" )

	minus2Hours := time.Duration(-2 * time.Hour)
	lastModified2HoursAgo := time.Now().UTC().Add(minus2Hours)
	expired = expiredKey(&s3.GetObjectOutput{
		LastModified: &lastModified2HoursAgo}, 
		int64(3600))
	assert.Equal(t, expired, true, "Expected true for an item modified 2 hours ago" )
}

func Test_expiredKey1(t *testing.T) {
	expire := time.Duration(-2 * time.Hour)
	lastModified2HoursAgo := time.Now().UTC().Add(expire)
	modifiedRightNow := time.Now().UTC()
	type args struct {
		obj           *s3.GetObjectOutput
		storageExpire int64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Modified 2 hours ago", 
		args : args {
			obj: &s3.GetObjectOutput{LastModified: &lastModified2HoursAgo}, 
			storageExpire: int64(3600),
		},
		want: true},
		{name: "Modified right now", 
			args : args {
				obj: &s3.GetObjectOutput{LastModified: &modifiedRightNow}, 
				storageExpire: int64(3600),
			},
			want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := expiredKey(tt.args.obj, tt.args.storageExpire); got != tt.want {
				t.Errorf("expiredKey() = %v, want %v", got, tt.want)
			}
		})
	}
}


func Test_delete(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("DeleteObject", mock.AnythingOfType("*s3.DeleteObjectInput")).Return(&s3.DeleteObjectOutput{}, nil)

	_, err := delete(svc, "bucket", "key")
	assert.Nil(t, err, "Expected no error")
}

func Test_deleteAll(t *testing.T) {
	svc := new(mocks.S3API)
	svc.On("ListObjects", mock.AnythingOfType("*s3.ListObjectsInput")).Return(&s3.ListObjectsOutput{
		Contents: []*s3.Object{
			&s3.Object{Key: aws.String("1")},
			&s3.Object{Key: aws.String("2")},
			&s3.Object{Key: aws.String("3")},
			},
		IsTruncated: aws.Bool(false),
	}, nil)
	svc.On("DeleteObjects", mock.AnythingOfType("*s3.DeleteObjectsInput")).Return(&s3.DeleteObjectsOutput{}, nil)

	err := deleteAll(svc, "bucket")
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