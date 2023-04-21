package ds

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var directoryIDRegex = regexp.MustCompile(`^d-[0-9a-f]{10}$`)

var directoryIDValidator validator.String = stringvalidator.RegexMatches(directoryIDRegex, "must be a valid Directory Service Directory ID")

var fqdnValidator validator.String = stringvalidator.RegexMatches(regexp.MustCompile(`^([a-zA-Z0-9]+[\\.-])+([a-zA-Z0-9])+[.]?$`), "must be a fully qualified domain name")

var trustPasswordValidator validator.String = stringvalidator.RegexMatches(regexp.MustCompile(`^(\p{L}|\p{Nd}|\p{P}| )+$`), "can contain upper- and lower-case letters, numbers, and punctuation characters")
