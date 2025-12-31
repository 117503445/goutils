package aliyun

import (
	"context"
	"fmt"

	cr20181201 "github.com/alibabacloud-go/cr-20181201/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AcrClientParams struct {
	Region string

	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
}

func NewAcrClient(ctx context.Context, params AcrClientParams) (*cr20181201.Client, error) {
	if params.AccessKeyId == "" || params.AccessKeySecret == "" {
		return nil, fmt.Errorf("access key id or access key secret is required")
	}
	if params.Region == "" {
		return nil, fmt.Errorf("region is required")
	}

	config := &openapi.Config{
		AccessKeyId:     tea.String(params.AccessKeyId),
		AccessKeySecret: tea.String(params.AccessKeySecret),
		SecurityToken:   tea.String(params.SecurityToken),
		Endpoint:        tea.String(fmt.Sprintf("cr.%s.aliyuncs.com", params.Region)),
	}
	return cr20181201.NewClient(config)
}
