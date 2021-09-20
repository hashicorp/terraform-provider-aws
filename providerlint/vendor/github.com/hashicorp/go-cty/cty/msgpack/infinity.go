package msgpack

import (
	"math"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var negativeInfinity = math.Inf(-1)
var positiveInfinity = math.Inf(1)
