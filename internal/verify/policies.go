package verify

import (
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func NewPolicyUnlessEquivalent(old, new string) (string, error) {
	// valid empty JSON is "{}" not "" so handle special case to avoid
	// Error unmarshaling policy: unexpected end of JSON input
	if old == "" || new == "" {
		return new, nil
	}

	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)

	if err != nil {
		return "", err
	}

	if equivalent {
		return old, nil
	}

	return new, nil
}
