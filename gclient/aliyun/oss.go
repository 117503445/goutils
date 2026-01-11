package aliyun

import (
	"context"
	"fmt"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

type OssClientParams struct {
	Region          string
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

func NewOssClient(ctx context.Context, params OssClientParams) (*oss.Client, error) {
	if params.AccessKeyId == "" || params.AccessKeySecret == "" {
		return nil, fmt.Errorf("access key id or access key secret is required")
	}
	if params.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	provider := credentials.NewStaticCredentialsProvider(params.AccessKeyId, params.AccessKeySecret, params.SecurityToken)

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(provider).WithRegion(params.Region)
	return oss.NewClient(cfg), nil
}
