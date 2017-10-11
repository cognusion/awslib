package awslib

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3PresignV4 "presigns" an S3 GET URL
func S3PresignV4(bucketName, filePath, bucketRegion string, expireFromNow time.Duration, creds *credentials.Credentials) (signedUrl string, err error) {

	if creds == nil {
		creds = AWSSession.Config.Credentials
	}

	if bucketRegion == "" {
		bucketRegion = *AWSSession.Config.Region
	}

	var url string
	if bucketRegion == "us-east-1" {
		// Legacy - omit region from URL
		url = fmt.Sprintf("https://%s.amazonaws.com/%s%s",
			"s3", bucketName, filePath)
	} else {
		url = fmt.Sprintf("https://%s.%s.amazonaws.com/%s%s",
			"s3", bucketRegion, bucketName, filePath)
	}

	req, _ := http.NewRequest("GET", url, nil)

	sign := v4.NewSigner(creds)
	sign.DisableURIPathEscaping = true
	sign.DisableRequestBodyOverwrite = true

	_, err = sign.Presign(req, nil, "s3", bucketRegion, expireFromNow, time.Now())
	if err == nil {
		signedUrl = req.URL.String()
	}

	return
}

// S3urlToParts explodes an s3://bucket/path/file url into its parts
func S3urlToParts(url string) (bucket, filePath, filename string) {

	// Trim the s3 URI prefix
	if strings.HasPrefix(url, "s3://") {
		url = strings.Replace(url, "s3://", "", 1)
	}

	// Extract the basename
	filename = filepath.Base(url)

	// Split the bucket from the path
	sparts := strings.SplitN(url, "/", 2)

	// Assuming everything goes well,
	// fill in the vars
	if len(sparts) == 2 {
		bucket = sparts[0]
		filePath = sparts[1]
	}

	return
}

// BucketToFile copies a file from an S3 bucket to a local file
func BucketToFile(bucket, bucketPath, filename string) (size int64, err error) {

	// Open the file
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(nil)
	size, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(bucketPath),
		})

	return
}

// BucketToFileVersion copies a version of a file from an S3 bucket to a local file
func BucketToFileVersion(bucket, bucketPath, filename, version string) (size int64, err error) {

	// Open the file
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(nil)
	size, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket:    aws.String(bucket),
			Key:       aws.String(bucketPath),
			VersionId: aws.String(version),
		})

	return
}

// FileToBucket copies a "local" file to an S3 bucket
func FileToBucket(filename, bucket string) (size int64, err error) {

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	// Get the filesize
	fi, ferr := file.Stat()
	if ferr == nil {
		size = fi.Size()
	}

	// Extract the basename
	baseFilename := filepath.Base(filename)

	// Setup the uploader, and git'r'done
	svc := s3manager.NewUploader(nil)
	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(baseFilename),
		Body:   file,
	})

	return
}

// Calculate_etag reads the specified filename in chunkSizeMB blocks, and returns
// the S3 multipart-upload etag/md5sum-of-sums or an error. The read capped and
// buffered to prevent heap allocations.
func Calculate_etag(filename string, chunkSizeMB int64) (etag string, err error) {

	var (
		count      = 0
		subtag     []byte
		readBuffer = make([]byte, 1048576*chunkSizeMB)
	)

	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	for {
		var size = -1
		size, err = file.Read(readBuffer)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return
			}
		}

		sum := md5.Sum(readBuffer[:size])
		subtag = append(subtag, sum[:]...)
		count += 1

		if size < len(readBuffer) {
			// Buffer wasn't filled. Tip.
			break
		}
	}

	etag = fmt.Sprintf("%x-%d", md5.Sum([]byte(subtag)), count)
	err = nil
	return
}
