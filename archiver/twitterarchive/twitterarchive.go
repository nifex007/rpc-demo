package twitterarchive

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/twitchtv/twirp"
)

// TwitterArchiveServer implements archiving for Tweets
type TwitterArchiveServer struct {
	KinesisService    *kinesis.Kinesis
	KinesisStreamName string
}

// New instantiates a new instance of the TwitterArchiveServer
func New(region string, streamName string) *TwitterArchiveServer {

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil
	}

	// Set the AWS Region that the service clients should use
	cfg.Region = region

	return &TwitterArchiveServer{
		KinesisService:    kinesis.New(cfg),
		KinesisStreamName: streamName,
	}

}

// Create receives a Tweet, and archives it to an Amazon Kinesis Stream.
func (t *TwitterArchiveServer) Create(ctx context.Context, req *CreateRequest) (*CreateResponse, error) {

	if len(req.Name) < 1 {
		return nil, twirp.InvalidArgumentError("name", "not set")
	}

	if len(req.Message) < 1 {
		return nil, twirp.InvalidArgumentError("message", "not set")
	}

	data := fmt.Sprintf(`{"name": "%s", "message": "%s"}`, req.Name, req.Message)
	putreq := t.KinesisService.PutRecordRequest(&kinesis.PutRecordInput{
		StreamName:   aws.String(t.KinesisStreamName),
		PartitionKey: aws.String(string(time.Now().Nanosecond())),
		Data:         []byte(data),
	})

	putresp, err := putreq.Send()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	fmt.Printf("(archived %s) %s: %s\n", *putresp.SequenceNumber, req.Name, req.Message)

	return &CreateResponse{
		Sequence: *putresp.SequenceNumber,
		Shard:    *putresp.ShardId,
	}, nil

}
