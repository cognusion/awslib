package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
)

// Ec2Info is a helper structure ...
type Ec2Info struct {
	State string
	Class string
	Arch  string
	Ltime string
	Az    string
}

//EbInfo is a helper structure...
type EbInfo struct {
	State    string
	Color    string
	Causes   []*string
	Degraded int64
	Info     int64
	NoData   int64
	Ok       int64
	Pending  int64
	Severe   int64
	Unknown  int64
	Warning  int64
}

// NewEbInfo returns an EbInfo struct or an error for the specified EB environment
func NewEbInfo(ebenv string, session *Session) (ebInfo *EbInfo, err error) {
	svc := elasticbeanstalk.New(session.AWS)
	input := &elasticbeanstalk.DescribeEnvironmentHealthInput{
		AttributeNames: []*string{
			aws.String("All"),
		},
		EnvironmentName: aws.String(ebenv),
	}

	result, err := svc.DescribeEnvironmentHealth(input)
	if err != nil {
		return nil, err
	}

	ebInfo = &EbInfo{
		State:    *result.HealthStatus,
		Color:    *result.Color,
		Causes:   result.Causes,
		Degraded: *result.InstancesHealth.Degraded,
		Info:     *result.InstancesHealth.Info,
		NoData:   *result.InstancesHealth.NoData,
		Ok:       *result.InstancesHealth.Ok,
		Pending:  *result.InstancesHealth.Pending,
		Severe:   *result.InstancesHealth.Severe,
		Unknown:  *result.InstancesHealth.Unknown,
		Warning:  *result.InstancesHealth.Warning,
	}

	return
}
