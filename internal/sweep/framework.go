package sweep

import (
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

// Terraform Plugin Framework variants of sweeper helpers.

type FrameworkSupplementalAttribute = framework.FrameworkSupplementalAttribute

type SweepFrameworkResource = framework.SweepFrameworkResource

var NewSweepFrameworkResource = framework.NewSweepFrameworkResource
