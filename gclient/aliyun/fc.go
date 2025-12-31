package aliyun

import (
	"context"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc20230330 "github.com/alibabacloud-go/fc-20230330/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

type Fc3ClientParams struct {
	Region    string
	AccountID string

	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

func NewFc3Client(ctx context.Context, params Fc3ClientParams) (*fc20230330.Client, error) {
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
		Endpoint:        tea.String(fmt.Sprintf("%s.%s.fc.aliyuncs.com", params.AccountID, params.Region)),
	}
	return fc20230330.NewClient(config)
}
