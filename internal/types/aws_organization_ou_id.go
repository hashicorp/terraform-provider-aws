package types

import "github.com/YakDriver/regexache"

// IsAWSOrganizationOUID returns whether or not the specified string is a valid AWS Organizational Unit ID.
func IsAWSOrganizationOUID(s string) bool {
	return regexache.MustCompile(`^ou-[0-9a-z]{4,32}-[a-z0-9]{8,32}$`).MatchString(s)
}
