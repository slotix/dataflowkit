package fetch

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/slotix/dataflowkit/errs"

	"github.com/slotix/dataflowkit/splash"
	"github.com/spf13/viper"
)

type s3Middleware struct {
	Service
}

// implement function to return ServiceMiddleware
func S3Middleware() ServiceMiddleware {
	return func(next Service) Service {
		return s3Middleware{next}
	}
}

func (mw s3Middleware) Fetch(req interface{}) (output interface{}, err error) {
	//init
	bucket := viper.GetString("FETCH_BUCKET")
	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region from the shared config file. (~/.aws/config).
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)
	var expErr error
	sReq := req.(splash.Request)
	url := sReq.GetURL()
	//if the item (req.URL) is in AWS S3 storage return local copy
	buf := &aws.WriteAtBuffer{}
	_, noSuchKeyErr := downloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(url),
		})

	if noSuchKeyErr == nil {
		var sResponse *splash.Response
		if err := json.Unmarshal(buf.Bytes(), &sResponse); err != nil {
			logger.Println("Json Unmarshall error", err)
		}
		//Error responses: a 404 (Not Found) may be cached.  
		if sResponse.Response.Status == 404 {
			return nil, &errs.NotFound{URL: url}
		}
		//check if item is expired.
	//	logger.Println(sResponse.Expires)
	//	logger.Println(time.Now().UTC())

		diff := sResponse.Expires.Sub(time.Now().UTC())
		logger.Printf("%s: cache lifespan is %+v\n", url, diff)
		//logger.Println(diff > 0)

		if diff > 0 { //if cached value is valid return it
			output = sResponse
			return output, nil
		}
		//otherwise cached item is expired and should be refetched
		expErr = &errs.ExpiredItemOrNotCacheable{}
	}

	
	//fetch results if there is nothing in a cache
	if noSuchKeyErr != nil {
		logger.Println(noSuchKeyErr.(s3.RequestFailure).Message())
	} else {
		logger.Println(expErr)
	}
	// if there is no cached copy of web page in S3 bucket, its content should be retrieved from the actual website. Current err value is not passed outside.
	err = nil
	resp, err := mw.Service.Fetch(req)
	if err != nil {
		return nil, err
	}
	if sResponse, ok := resp.(*splash.Response); ok {
		logger.Println("Cachable? ", sResponse.Cacheable)
		response, err := json.Marshal(resp)
		if err != nil {
			logger.Printf(err.Error())
		}
		r := bytes.NewReader(response)
		_, err = uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(url),
			Body:   r,
			//Expires metadata is not used here as it duplicates splash.Response Expires value.
			//Expires: &sResponse.Expires,
		})
		if err != nil {
			logger.Println(err)
		}
		
		output = sResponse
	}
	return
}
