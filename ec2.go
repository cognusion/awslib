package awslib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
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

// NewEc2Info returns an Ec2Info struct or an error from an instance ID
// Assumes InitAWS has been called.
func NewEc2Info(instance string) (iInfo *Ec2Info, err error) {

	svc := ec2.New(AWSSession)
	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instance),
		},
	}
	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return
	}
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if inst.Platform == nil || *inst.Platform != "windows" {
				// Not Windows, phew
				//awsconf.AddHost(NewHostFromInstance(inst))
			}
		}
	}
	/*
		iInfo = &Ec2Info{
			State: ,
			Class: ,
			Arch: resp.,
			Ltime: ,
			Az: ,
		}
	*/

	return
}

// ELB_HostCounts returns the last healthy and unhealthy -hostcounts.
// Assumes InitAWS has been called.
func ELB_HostCounts(instance string) (healthyPoint, unhealthyPoint *cloudwatch.Datapoint, err error) {

	Uresp, err := getMetrics("LoadBalancerName", instance, "AWS/ELB", "UnHealthyHostCount", "Maximum", "Count")
	if err != nil {
		return
	}
	Hresp, err := getMetrics("LoadBalancerName", instance, "AWS/ELB", "HealthyHostCount", "Maximum", "Count")
	if err != nil {
		return
	}

	healthyPoint = lastMetric(Hresp)
	unhealthyPoint = lastMetric(Uresp)
	return
}

// NewEbInfo returns an EbInfo struct or an error for the specified EB environment
// Assumes InitAWS has been called.
func NewEbInfo(ebenv string) (ebInfo *EbInfo, err error) {
	svc := elasticbeanstalk.New(AWSSession)
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
