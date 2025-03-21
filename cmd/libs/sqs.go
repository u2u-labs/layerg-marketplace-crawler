package libs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var (
	QUEUE_URL  = ""
	AWS_REGION = ""
	SqsClient  *sqs.Client
)

func InitAWSConfig() aws.Config {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(AWS_REGION))
	if err != nil {
		panic(err)
	}
	return cfg
}

func InitSQSClient(cfg aws.Config) *sqs.Client {
	SqsClient = sqs.NewFromConfig(cfg)
	return SqsClient
}
