package s3

import (
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/spf13/viper"
)

var s *session.Session
var uploader *s3manager.Uploader

// Upload file to s3 bucket
func Upload(key string, file io.ReadSeeker) (string, *response.Error) {
	bucketName := viper.GetString("s3_bucket_name")

	obInput := &s3manager.UploadInput{
		Body:   file,
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	result, err := uploader.Upload(obInput)
	if err != nil {
		var errMsg string
		if awsErr, ok := err.(awserr.Error); ok {
			errMsg = awsErr.Error()

			if origErr := awsErr.OrigErr(); origErr != nil {
				fmt.Printf("original error: %v", origErr)
			}
		}
		return "", response.NewErrorf(
			http.StatusInternalServerError,
			"error uploading file to s3 bucket: %v",
			errMsg,
		)
	}

	return result.Location, nil
}

// InitSession set up s3 session and service for file uploading
func InitSession() *response.Error {
	region := viper.GetString("s3_region")
	id := viper.GetString("s3_id")
	secret := viper.GetString("s3_secret")

	var err error
	s, err = session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
	if err != nil {
		return response.NewErrorf(
			http.StatusInternalServerError,
			"error creating aws session: %v",
			err,
		)
	}

	uploader = s3manager.NewUploader(s)

	return nil
}
