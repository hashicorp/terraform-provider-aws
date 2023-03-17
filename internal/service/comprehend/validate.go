package comprehend

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const (
	modelIdentifierMaxLen       = 63 // Documentation says 256, Console says 63
	modelIdentifierPrefixMaxLen = modelIdentifierMaxLen - resource.UniqueIDSuffixLength
)

var validModelName = validIdentifier
var validModelVersionName = validation.Any( // nosemgrep:ci.avoid-string-is-empty-validation
	validation.StringIsEmpty,
	validIdentifier,
)
var validModelVersionNamePrefix = validIdentifierPrefix

var validIdentifier = validation.All(
	validation.StringLenBetween(1, modelIdentifierMaxLen),
	validIdentifierPattern,
)

var validIdentifierPrefix = validation.All(
	validation.StringLenBetween(1, modelIdentifierPrefixMaxLen),
	validIdentifierPattern,
)

var validIdentifierPattern = validation.StringMatch(regexp.MustCompile(`^[[:alnum:]-]+$`), "must contain A-Z, a-z, 0-9, and hypen (-)")
