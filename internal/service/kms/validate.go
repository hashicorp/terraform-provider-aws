package kms

import (
	"fmt"
	"regexp"
)

func validGrantName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if len(value) > 256 {
		es = append(es, fmt.Errorf("%s can not be greater than 256 characters", k))
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9:/_-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%s must only contain [a-zA-Z0-9:/_-]", k))
	}

	return
}

func validNameForDataSource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !regexp.MustCompile(`^(alias/)[a-zA-Z0-9/_-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [a-zA-Z0-9/_-]", k))
	}
	return
}

func validNameForResource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if regexp.MustCompile(`^(alias/aws/)`).MatchString(value) {
		es = append(es, fmt.Errorf("%q cannot begin with reserved AWS CMK prefix 'alias/aws/'", k))
	}

	if !regexp.MustCompile(`^(alias/)[a-zA-Z0-9/_-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [a-zA-Z0-9/_-]", k))
	}
	return
}

func validKey(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	arnPrefixPattern := `arn:[^:]+:kms:[^:]+:[^:]+:`
	keyIdPattern := "[A-Za-z0-9-]+"
	keyArnPattern := arnPrefixPattern + "key/" + keyIdPattern
	aliasNamePattern := "alias/[a-zA-Z0-9:/_-]+"
	aliasArnPattern := arnPrefixPattern + aliasNamePattern
	if !regexp.MustCompile(fmt.Sprintf("^%s$", keyIdPattern)).MatchString(value) &&
		!regexp.MustCompile(fmt.Sprintf("^%s$", keyArnPattern)).MatchString(value) &&
		!regexp.MustCompile(fmt.Sprintf("^%s$", aliasNamePattern)).MatchString(value) &&
		!regexp.MustCompile(fmt.Sprintf("^%s$", aliasArnPattern)).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q must be one of the following patterns: %s, %s, %s or %s", k, keyIdPattern, keyArnPattern, aliasNamePattern, aliasArnPattern))
	}
	return
}
