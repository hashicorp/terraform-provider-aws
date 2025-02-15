package models

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type LocationInfoModel struct {
	Name types.String                              `tfsdk:"name"`
	Type fwtypes.StringEnum[awstypes.LocationType] `tfsdk:"type"`
}
