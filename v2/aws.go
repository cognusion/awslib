package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cognusion/go-timings"

	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	// DebugOut is a log.Logger for debug messages
	DebugOut = log.New(io.Discard, "", 0)
	// TimingOut is a log.Logger for timing-related debug messages. DEPRECATED
	TimingOut = log.New(io.Discard, "[TIMING] ", 0)
)

// Session is a container around an AWS Session, to make AWS operations easier
type Session struct {
	// AWS is the raw, hopefully initialized AWS Session
	AWS *session.Session
	Me  *ec2metadata.EC2InstanceIdentityDocument
}

// NewSession returns a Session or an error
func NewSession(awsRegion, awsAccessKey, awsSecretKey string) (*Session, error) {

	s := Session{}
	awsSession, err := InitAWS(awsRegion, awsAccessKey, awsSecretKey)

	if err != nil {
		// Error initing session
		return nil, err
	}
	s.AWS = awsSession

	idd, err := s.getMe()
	if err != nil {
		// Error getting ec2metadata
		return nil, err
	}
	s.Me = &idd
	DebugOut.Printf("EC2 Me: %+v\n", idd)
	return &s, nil
}

// InitAWS optionally takes a region, accesskey and secret key,
// setting AWSSession to the resulting session. If values aren't
// provided, the well-known environment variables (WKE) are
// consulted. If they're not available, and running in an EC2
// instance, then it will use the local IAM role
func InitAWS(awsRegion, awsAccessKey, awsSecretKey string) (*session.Session, error) {

	config := aws.NewConfig()

	// Region
	if awsRegion != "" {
		// CLI trumps
		config.Region = aws.String(awsRegion)
	} else if os.Getenv("AWS_DEFAULT_REGION") != "" {
		// Env is good, too
		config.Region = aws.String(os.Getenv("AWS_DEFAULT_REGION"))
	} else {
		// Grab it from this EC2 instance, maybe
		region, err := GetAwsRegionE()
		if err != nil {
			return nil, fmt.Errorf("cannot set AWS region: '%w'", err)
		}
		config.Region = aws.String(region)
	}

	// Creds
	if awsAccessKey != "" && awsSecretKey != "" {
		// CLI trumps
		creds := credentials.NewStaticCredentials(
			awsAccessKey,
			awsSecretKey,
			"")
		config.Credentials = creds
	} else if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		// Env is good, too
		creds := credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"")
		config.Credentials = creds
	}

	return session.NewSession(config)
}

// BucketToFile copies a file from an S3 bucket to a local file
func (s *Session) BucketToFile(bucket, bucketPath, filename string) (size int64, err error) {

	// Open the file
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return
	}
	defer file.Close()

	downloader := s3manager.NewDownloader(s.AWS)
	size, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(bucketPath),
		})

	return
}

// getMe grabs the InstanceIdentityDocument for the running instance
func (s *Session) getMe() (ec2metadata.EC2InstanceIdentityDocument, error) {
	// {DevpayProductCodes:[] AvailabilityZone:us-east-1d PrivateIP:10.2.21.50 Version:2017-09-30
	//  Region:us-east-1 InstanceID:i-032681bb83e1de5cf BillingProducts:[] InstanceType:t2.medium
	//  AccountID:929091317894 PendingTime:2018-06-27 18:06:50 +0000 UTC ImageID:ami-55ef662f KernelID:
	//  RamdiskID: Architecture:x86_64}
	return ec2metadata.New(s.AWS).GetInstanceIdentityDocument()
}

// GetInstanceAZByIP returns an Availability Zone or an error
func (s *Session) GetInstanceAZByIP(ip string) (string, error) {

	// Timings
	t := timings.Tracker{}
	t.Start()
	defer TimingOut.Printf("GetInstanceAZByIP took %s\n", t.Since().String())

	F := ec2.Filter{
		Name:   aws.String("private-ip-address"),
		Values: []*string{&ip},
	}
	DII := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{&F},
	}

	svc := ec2.New(s.AWS)

	var (
		result *ec2.DescribeInstancesOutput
		err    error
	)

	result, err = svc.DescribeInstances(&DII)
	if err != nil {
		return "", err
	}
	for _, res := range result.Reservations {
		for _, ins := range res.Instances {
			return *ins.Placement.AvailabilityZone, nil
		}
	}

	return "", nil
}

// GetInstancesAZByIP returns a map of IPs to Availability Zones or an error
func (s *Session) GetInstancesAZByIP(ips []*string) (*map[string]string, error) {

	var (
		mss   = make(map[string]string)
		token *string
	)

	F := ec2.Filter{
		Name:   aws.String("private-ip-address"),
		Values: ips,
	}
	DII := ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{&F},
	}

	svc := ec2.New(s.AWS)

	token = aws.String("nil") // just because
	for token != nil {
		var (
			result *ec2.DescribeInstancesOutput
			err    error
		)

		if *token == "nil" {
			// First run
			result, err = svc.DescribeInstances(&DII)
		} else {
			// Pages
			DIIP := ec2.DescribeInstancesInput{
				NextToken: token,
			}
			result, err = svc.DescribeInstances(&DIIP)
		}
		if err != nil {
			return nil, err
		}
		for _, res := range result.Reservations {
			for _, ins := range res.Instances {
				mss[*ins.PrivateIpAddress] = *ins.Placement.AvailabilityZone
			}
		}

		token = result.NextToken
	}

	return &mss, nil
}

// GetAwsRegion returns the region as a string,
// first consulting the well-known environment variables,
// then falling back EC2 metadata calls
func GetAwsRegion() (region string) {
	region, _ = GetAwsRegionE()
	return
}

// GetAwsRegionE returns the region as a string and and error,
// first consulting the well-known environment variables,
// then falling back EC2 metadata calls
func GetAwsRegionE() (region string, err error) {

	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	} else {
		// Grab it from this EC2 instace
		region, err = ec2metadata.New(session.New()).Region()
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
