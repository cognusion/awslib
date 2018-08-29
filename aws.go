package awslib

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

// AWSSession is a global variable holding an
// AWS session.Session. Call InitAWS to set
var AWSSession *session.Session

// InitAWS optionally takes a region, accesskey and secret key,
// setting AWSSession to the resulting session. If values aren't
// provided, the well-known environment variables (WKE) are
// consulted. If they're not available, and running in an EC2
// instance, then it will use the local IAM role
func InitAWS(awsRegion, awsAccessKey, awsSecretKey string) {

	AWSSession = session.New()

	// Region
	if awsRegion != "" {
		// CLI trumps
		AWSSession.Config.Region = aws.String(awsRegion)
	} else if os.Getenv("AWS_DEFAULT_REGION") != "" {
		// Env is good, too
		AWSSession.Config.Region = aws.String(os.Getenv("AWS_DEFAULT_REGION"))
	} else {
		// Grab it from this EC2 instance, maybe
		region, err := getAwsRegionE()
		if err != nil {
			fmt.Printf("Cannot set AWS region: '%v'\n", err)
			os.Exit(1)
		}
		AWSSession.Config.Region = aws.String(region)
	}

	// Creds
	if awsAccessKey != "" && awsSecretKey != "" {
		// CLI trumps
		creds := credentials.NewStaticCredentials(
			awsAccessKey,
			awsSecretKey,
			"")
		AWSSession.Config.Credentials = creds
	} else if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		// Env is good, too
		creds := credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"")
		AWSSession.Config.Credentials = creds
	}

}

// getAwsRegion returns the region as a string,
// first consulting the well-known environment variables,
// then falling back EC2 metadata calls
func getAwsRegion() (region string) {
	region, _ = getAwsRegionE()
	return
}

// getAwsRegionE returns the region as a string and and error,
// first consulting the well-known environment variables,
// then falling back EC2 metadata calls
func getAwsRegionE() (region string, err error) {

	if os.Getenv("AWS_DEFAULT_REGION") != "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	} else {
		// Grab it from this EC2 instace
		region, err = ec2metadata.New(session.New()).Region()
	}
	return
}

// listMetrics ...
func listMetrics(dimensionName, dimensionValue, namespace string) (resp *cloudwatch.ListMetricsOutput, err error) {
	svc := cloudwatch.New(nil)
	params := &cloudwatch.ListMetricsInput{
		Dimensions: []*cloudwatch.DimensionFilter{
			{
				Name:  aws.String(dimensionName), // Required
				Value: aws.String(dimensionValue),
			},
		},
		Namespace: aws.String(namespace),
	}
	resp, err = svc.ListMetrics(params)

	return
}

// lastMetric ...
func lastMetric(metrics *cloudwatch.GetMetricStatisticsOutput) (point *cloudwatch.Datapoint) {

	newest := time.Now().Add(-5 * time.Hour)

	for _, p := range metrics.Datapoints {
		if p.Timestamp.After(newest) {
			newest = *p.Timestamp
			point = p
		}
	}

	return
}

func GetLastCloudWatchValue(dimensionName, dimensionValue, namespace, metric, stat, unit string) (point *cloudwatch.Datapoint, err error) {
	var resp *cloudwatch.GetMetricStatisticsOutput
	resp, err = getMetrics(dimensionName, dimensionValue, namespace, metric, stat, unit)
	if err != nil {
		return
	}
	point = lastMetric(resp)
	return
}

// getMetrics ...
func getMetrics(dimensionName, dimensionValue, namespace, metric, stat, unit string) (resp *cloudwatch.GetMetricStatisticsOutput, err error) {
	svc := cloudwatch.New(AWSSession)

	now := time.Now()
	params := &cloudwatch.GetMetricStatisticsInput{
		EndTime:    aws.Time(now),                       // Required
		MetricName: aws.String(metric),                  // Required
		Namespace:  aws.String(namespace),               // Required
		Period:     aws.Int64(60),                       // Required
		StartTime:  aws.Time(now.Add(-5 * time.Minute)), // Required
		Statistics: []*string{ // Required
			aws.String(stat), // Required
		},
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String(dimensionName), // Required
				Value: aws.String(dimensionValue),
			},
		},
		Unit: aws.String(unit),
	}
	resp, err = svc.GetMetricStatistics(params)

	return
}

// nameToResourceType ...
func nameToResourceType(name string) (resourceType string) {
	// EC2 i-
	// RDS .rds.amazonaws.com
	// ELB
	// EBS vol-
	// EIP eipalloc-
	return
}
