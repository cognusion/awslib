package awslib

import (
	"github.com/spf13/cast"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
)

// RDSStorageInfo ...
type RDSStorageInfo struct {
	Allocated float64
	Free      float64
	Used      float64
	PercFree  float64
	PercUsed  float64
	ReadIops  float64
	WriteIops float64
}

// NewRDSStorageInfo returns an RDSStorageInfo struct
// Assumes InitAWS has been called.
func NewRDSStorageInfo(instance string) (sInfo *RDSStorageInfo, err error) {

	iInfo, err := RDS_Instance(instance)
	if err != nil {
		return
	}

	fInfo, err := RDS_FreeStorageSpace(instance)
	if err != nil {
		return
	}

	wInfo, err := RDS_WriteIOPS(instance)
	if err != nil {
		return
	}

	rInfo, err := RDS_ReadIOPS(instance)
	if err != nil {
		return
	}

	storage := cast.ToFloat64(iInfo.AllocatedStorage) * 1073741824
	freeStorage := *fInfo.Maximum
	freePerc := (freeStorage / storage) * 100

	sInfo = &RDSStorageInfo{
		Allocated: storage,
		Free:      freeStorage,
		PercFree:  freePerc,
		Used:      storage - freeStorage,
		PercUsed:  100 - freePerc,
		ReadIops:  *rInfo.Maximum,
		WriteIops: *wInfo.Maximum,
	}

	return
}

// RDS_Instance ...
// Assumes InitAWS has been called.
func RDS_Instance(instance string) (i *rds.DBInstance, err error) {

	svc := rds.New(AWSSession)

	params := &rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(instance),
	}
	resp, err := svc.DescribeDBInstances(params)
	if err != nil {
		return
	}

	i = resp.DBInstances[0]
	return
}

// RDS_CPUUtilization returns the last CPUUtilization datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_CPUUtilization(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "CPUUtilization", "Maximum", "Percent")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}

// RDS_DatabaseConnections returns the last DatabaseConnections datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_DatabaseConnections(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "DatabaseConnections", "Maximum", "Count")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}

// RDS_FreeableMemory returns the last FreeableMemory datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_FreeableMemory(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "FreeableMemory", "Maximum", "Bytes")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}

// RDS_FreeStorageSpace returns the last FreeStorageSpace datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_FreeStorageSpace(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "FreeStorageSpace", "Maximum", "Bytes")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}

// RDS_ReadIOPS returns the last ReadIOPS datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_ReadIOPS(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "ReadIOPS", "Maximum", "Count/Second")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}

// RDS_WriteIOPS returns the last WriteIOPS datapoint for the specified instance.
// Assumes InitAWS has been called.
func RDS_WriteIOPS(instance string) (point *cloudwatch.Datapoint, err error) {

	resp, err := getMetrics("DBInstanceIdentifier", instance, "AWS/RDS", "WriteIOPS", "Maximum", "Count/Second")
	if err != nil {
		return
	}

	point = lastMetric(resp)
	return
}
