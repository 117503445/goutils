package aliyun

import (
	"context"
	"fmt"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

type SlsClientParams struct {
	Region string

	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

func NewSlsClient(ctx context.Context, params SlsClientParams) (sls.ClientInterface, error) {
	if params.AccessKeyId == "" || params.AccessKeySecret == "" {
		return nil, fmt.Errorf("access key id or access key secret is required")
	}
	if params.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	var provider sls.CredentialsProvider

	if params.SecurityToken == "" {
		provider = sls.NewStaticCredentialsProvider(params.AccessKeyId, params.AccessKeySecret, "")
	} else {
		provider = sls.NewStaticCredentialsProvider(params.AccessKeyId, params.AccessKeySecret, params.SecurityToken)
	}
	endpoint := params.Region + ".log.aliyuncs.com"
	client := sls.CreateNormalInterfaceV2(endpoint, provider)
	client.SetAuthVersion(sls.AuthV4)
	client.SetRegion(params.Region)
	return client, nil
}
