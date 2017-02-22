package awslib

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Ec2Info is a helper structure ...
type Ec2Info struct {
	State string
	Class string
	Arch  string
	Ltime string
	Az    string
}

// NewEc2Info returns an Ec2Info struct from an instance ID
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
