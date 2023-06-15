package sweep

import (
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

type SweepResource = sdk.SweepResource

var NewSweepResource = sdk.NewSweepResource

var DeleteResource = sdk.DeleteResource

var ReadResource = sdk.ReadResource
