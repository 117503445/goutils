package aliyun

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	ecs20140526 "github.com/alibabacloud-go/ecs-20140526/v7/client"
	"github.com/alibabacloud-go/tea/tea"
)

type EcsClientParams struct {
	Region    string
	AccountID string

	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

func NewEcsClient(ctx context.Context, params EcsClientParams) (*ecs20140526.Client, error) {
	if params.AccessKeyId == "" || params.AccessKeySecret == "" {
		return nil, fmt.Errorf("access key id or access key secret is required")
	}
	if params.Region == "" {
		return nil, fmt.Errorf("region is required")
	}
	if params.AccountID == "" {
		return nil, fmt.Errorf("account id is required")
	}
	config := &openapi.Config{
		AccessKeyId:     tea.String(params.AccessKeyId),
		AccessKeySecret: tea.String(params.AccessKeySecret),
		SecurityToken:   tea.String(params.SecurityToken),
		Endpoint:        tea.String(fmt.Sprintf("ecs.%s.aliyuncs.com", params.Region)),
	}
	return ecs20140526.NewClient(config)
}
