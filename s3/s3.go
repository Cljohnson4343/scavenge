package s3

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/spf13/viper"
)

var s *session.Session
var uploader *s3manager.Uploader
var bucketName string

// Upload file to s3 bucket
func Upload(key string, file io.ReadSeeker) (string, *response.Error) {

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

// GetKey returns the s3 object key from an objects url
func GetKey(url string) string {
	s := strings.Split(url, "/")
	return s[len(s)-1]
}

// Delete deletes an object from the s3 media bucket
func Delete(key string) *response.Error {
	svc := s3.New(s)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	result, err := svc.DeleteObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return response.NewErrorf(
			http.StatusBadRequest,
			"error deleting s3 object: %v",
			err,
		)
	}

	fmt.Println(result)
	return nil
}

// InitSession set up s3 session and service for file uploading
func InitSession() *response.Error {
	region := viper.GetString("s3_region")
	id := viper.GetString("s3_id")
	secret := viper.GetString("s3_secret")
	bucketName = viper.GetString("s3_bucket_name")

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
