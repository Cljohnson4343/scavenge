package s3

import (
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/spf13/viper"
)

var s *session.Session
var svc *s3.S3

func Upload(key string, contentType string, size int64, file io.ReadSeeker) (*s3.PutObjectOutput, *response.Error) {
	bucketName := viper.GetString("s3_bucket_name")

	obInput := &s3.PutObjectInput{
		ACL:                  aws.String("public"),
		Body:                 file,
		Bucket:               aws.String(bucketName),
		ContentLength:        aws.Int64(size),
		ContentType:          aws.String(contentType),
		Key:                  aws.String(key),
		ServerSideEncryption: aws.String("AES256"),
	}

	result, err := svc.PutObject(obInput)
	if err != nil {
		return nil, response.NewErrorf(
			http.StatusInternalServerError,
			"error uploading file to s3 bucket: %v",
			err,
		)
	}

	return result, nil
}

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

	svc = s3.New(s)

	return nil
}
