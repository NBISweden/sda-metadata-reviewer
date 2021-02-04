package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

type s3Backend struct {
	Client   *s3.S3
	Uploader *s3manager.Uploader
	Bucket   string
	Conf     *S3Config
}

// S3Config stores information about the S3 storage backend
type S3Config struct {
	URL               string
	Port              int
	AccessKey         string
	SecretKey         string
	Bucket            string
	Region            string
	UploadConcurrency int
	Chunksize         int
	Cacert            string
	NonExistRetryTime time.Duration
}

func newS3Backend(config S3Config) (*s3Backend, error) {
	s3Transport := transportConfigS3(config)
	client := http.Client{Transport: s3Transport}
	s3Session := session.Must(session.NewSession(
		&aws.Config{
			Endpoint:         aws.String(fmt.Sprintf("%s:%d", config.URL, config.Port)),
			Region:           aws.String(config.Region),
			HTTPClient:       &client,
			S3ForcePathStyle: aws.Bool(true),
			DisableSSL:       aws.Bool(strings.HasPrefix(config.URL, "http:")),
			Credentials:      credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, ""),
		},
	))

	// Attempt to create a bucket, but we really expect an error here
	// (BucketAlreadyOwnedByYou)
	_, err := s3.New(s3Session).CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(config.Bucket),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {

			if aerr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou &&
				aerr.Code() != s3.ErrCodeBucketAlreadyExists {
				log.Error("Unexpected issue while creating bucket", err)
			}
		}
	}

	sb := &s3Backend{
		Bucket: config.Bucket,
		Uploader: s3manager.NewUploader(s3Session, func(u *s3manager.Uploader) {
			u.PartSize = int64(config.Chunksize)
			u.Concurrency = config.UploadConcurrency
			u.LeavePartsOnError = false
		}),
		Client: s3.New(s3Session),
		Conf:   &config}

	_, err = sb.Client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &config.Bucket})

	if err != nil {
		return nil, err
	}

	return sb, nil
}

// transportConfigS3 is a helper method to setup TLS for the S3 client.
func transportConfigS3(config S3Config) http.RoundTripper {
	cfg := new(tls.Config)

	// Enforce TLS1.2 or higher
	cfg.MinVersion = 2

	// Read system CAs
	var systemCAs, _ = x509.SystemCertPool()
	if reflect.DeepEqual(systemCAs, x509.NewCertPool()) {
		log.Debug("creating new CApool")
		systemCAs = x509.NewCertPool()
	}
	cfg.RootCAs = systemCAs

	if config.Cacert != "" {
		cacert, e := ioutil.ReadFile(config.Cacert) // #nosec this file comes from our config
		if e != nil {
			log.Fatalf("failed to append %q to RootCAs: %v", cacert, e)
		}
		if ok := cfg.RootCAs.AppendCertsFromPEM(cacert); !ok {
			log.Debug("no certs appended, using system certs only")
		}
	}

	var trConfig http.RoundTripper = &http.Transport{
		TLSClientConfig:   cfg,
		ForceAttemptHTTP2: true}

	return trConfig
}

// GetFileSize returns the size of a specific object
func (sb *s3Backend) GetFileSize(filePath string) (bool, error) {
	if sb == nil {
		return false, fmt.Errorf("Invalid s3Backend")
	}

	_, err := sb.Client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(sb.Bucket),
		Key:    aws.String(filePath)})

	/*start := time.Now()

	retryTime := 2 * time.Minute
	if sb.Conf != nil {
		retryTime = sb.Conf.NonExistRetryTime
	}*/

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false, nil
			default:
				log.Info(err)
			}
		}
	}
	return true, nil

	// Retry on error up to five minutes to allow for
	// "slow writes' or s3 eventual consistency
	/*for err != nil && time.Since(start) < retryTime {
		r, err = sb.Client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(sb.Bucket),
			Key:    aws.String(filePath)})

		time.Sleep(1 * time.Second)

	}

	if err != nil {
		log.Errorln(err)
		return 0, err
	}*/

	//return *r.ContentLength, nil
}
