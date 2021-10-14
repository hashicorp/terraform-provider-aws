package cognitoidp

import (
	"fmt"
	"regexp"
)

func validResourceServerScopeName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q cannot be less than 1 character", k))
	}
	if len(value) > 256 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 256 character", k))
	}
	if !regexp.MustCompile(`[\x21\x23-\x2E\x30-\x5B\x5D-\x7E]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(`%q must satisfy regular expression pattern: [\x21\x23-\x2E\x30-\x5B\x5D-\x7E]+`, k))
	}
	return
}

func validUserGroupName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 128 {
		es = append(es, fmt.Errorf("%q cannot be longer than 128 character", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+`, k))
	}
	return
}

func validUserPoolEmailVerificationMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolEmailVerificationSubject(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf("%q can be composed of any kind of letter, symbols, numeric character, punctuation and whitespaces", k))
	}
	return
}

func validUserPoolID(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexp.MustCompile(`^[\w-]+_[0-9a-zA-Z]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be the region name followed by an underscore and then alphanumeric pattern", k))
	}
	return
}

func validUserPoolInviteTemplateEmailMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}

	if !regexp.MustCompile(`.*\{username\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {username}", k))
	}
	return
}

func validUserPoolInviteTemplateSMSMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}

	if !regexp.MustCompile(`.*\{username\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {username}", k))
	}
	return
}

func validUserPoolSchemaName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 20 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20 character", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+`, k))
	}
	return
}

func validUserPoolSMSAuthenticationMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolSMSVerificationMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolTemplateEmailMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolTemplateEmailMessageByLink(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{##[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*##\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{##[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*##\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`, k))
	}
	return
}

func validUserPoolTemplateEmailSubject(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`, k))
	}
	return
}

func validUserPoolTemplateEmailSubjectByLink(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`, k))
	}
	return
}

func validUserPoolTemplateSMSMessage(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 characters", k))
	}

	if len(value) > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 characters", k))
	}

	if !regexp.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}
