package awslib

import (
	"github.com/spf13/cast"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQS_GetQueueUrl returns the queue URL for the specified queue.
// Assumes InitAWS has been called.
func SQS_GetQueueUrl(queue string) (qurl string, err error) {
	svc := sqs.New(AWSSession)

	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queue),
	}
	resp, err := svc.GetQueueUrl(params)

	if err != nil {
		return
	}

	qurl = *resp.QueueUrl
	return
}

// SQS_Attributes returns an SqsInfo struct for the specified queue
// Assumes InitAWS has been called.
func SQS_Attributes(queue string) (sInfo *SqsInfo, err error) {
	sInfo, err = NewSqsInfo(queue)
	sInfo.sqs = nil // we want to kill that reference
	return
}

// SqsInfo ...
type SqsInfo struct {
	Messages          int64
	MessagesDelayed   int64
	MessagesInvisible int64
	ModifiedStamp     int64
	sqs               *sqs.SQS
	qurl              string
	lasterr           error
}

// NewSqsInfo ...
func NewSqsInfo(queue string) (s *SqsInfo, err error) {
	s.sqs = sqs.New(AWSSession)

	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(queue),
	}
	resp, err := s.sqs.GetQueueUrl(params)
	if err != nil {
		return
	}
	s.qurl = *resp.QueueUrl

	s.Refresh()
	err = s.LastError()

	return
}

// Refresh ...
func (s *SqsInfo) Refresh() {
	params := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.qurl), // Required
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
			aws.String("ApproximateNumberOfMessagesDelayed"),
			aws.String("ApproximateNumberOfMessagesNotVisible"),
			//aws.String("LastModifiedTimestamp"),
			//aws.String("FifoQueue"),
			//aws.String("ContentBasedDeduplication"),
		},
	}
	resp, err := s.sqs.GetQueueAttributes(params)
	if err != nil {
		s.lasterr = err
		return
	}
	s.lasterr = nil

	s.Messages = cast.ToInt64(resp.Attributes["ApproximateNumberOfMessages"])
	s.MessagesDelayed = cast.ToInt64(resp.Attributes["ApproximateNumberOfMessagesDelayed"])
	s.MessagesInvisible = cast.ToInt64(resp.Attributes["ApproximateNumberOfMessagesNotVisible"])
	//s.ModifiedStamp = cast.ToInt64(resp.Attributes["LastModifiedTimestamp"]),
}

// LastError returns the last polling error
func (s *SqsInfo) LastError() error {
	return s.lasterr
}
