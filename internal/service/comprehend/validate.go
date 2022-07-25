package comprehend

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	modelIdentifierMaxLen = 63 // Documentation says 256, Console says 63
)

var validModelName = validIdentifier
var validModelVersionName = validIdentifier

var validIdentifier = validation.All(
	validation.StringLenBetween(1, modelIdentifierMaxLen),
	validation.StringMatch(regexp.MustCompile(`[[:alnum:]-]`), "must contain A-Z, a-z, 0-9, and hypen (-)"),
)
