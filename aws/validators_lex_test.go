package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
)

func TestValidateLexNameTooShort(t *testing.T) {
	_, errs := validateLexName("", "name")
	if len(errs) == 0 {
		t.Fatalf("an empty string should be an invalid Lex name")
	}
}

func TestValidateLexNameTooLong(t *testing.T) {
	_, errs := validateLexName(acctest.RandString(lexNameMaxLength+1), "name")
	if len(errs) == 0 {
		t.Fatalf("a %d character string should be an invalid Lex name", lexNameMaxLength)
	}
}

func TestValidateLexNameInvalidCharacters(t *testing.T) {
	invalidNames := []string{
		"1test",
		"test-one",
		"test.one",
		"test1",
		"test1one",
		"_test",
		"test:one",
	}

	for _, invalidName := range invalidNames {
		_, errs := validateLexName(invalidName, "name")
		if len(errs) == 0 {
			t.Fatalf("%q should be an invalid Lex name", invalidName)
		}
	}
}

func TestValidateLexNameValid(t *testing.T) {
	validNames := []string{
		"test",
		"test_one",
		"test_one_",
		"Test",
		"TestOne",
		"Test_One",
		"Test_One_",
		"TEST",
	}

	for _, validName := range validNames {
		_, errs := validateLexName(validName, "name")
		if len(errs) > 0 {
			t.Fatalf("%q should be a valid Lex name, got %v", validName, errs)
		}
	}
}

func TestValidateLexVersionTooShort(t *testing.T) {
	_, errs := validateLexVersion("", "version")
	if len(errs) == 0 {
		t.Fatalf("an empty string should be an invalid Lex version")
	}
}

func TestValidateLexVersionTooLong(t *testing.T) {
	version := acctest.RandStringFromCharSet(lexVersionMaxLength+1, "0123456789")

	_, errs := validateLexVersion(version, "version")
	if len(errs) == 0 {
		t.Fatalf("%q should be an invalid Lex version", version)
	}
}

func TestValidateLexVersionInvalidCharacters(t *testing.T) {
	invalidVersions := []string{
		"one",
		"LATEST",
	}

	for _, invalidVersion := range invalidVersions {
		_, errs := validateLexVersion(invalidVersion, "version")
		if len(errs) == 0 {
			t.Fatalf("%q should be an invalid Lex version", invalidVersion)
		}
	}
}

func TestValidateLexVersionValid(t *testing.T) {
	validVersions := []string{
		"0",
		"1",
		"11",
		"$LATEST",
	}

	for _, validVersion := range validVersions {
		_, errs := validateLexVersion(validVersion, "version")
		if len(errs) > 0 {
			t.Fatalf("%q should be a valid Lex version, got %v", validVersion, errs)
		}
	}
}

func TestValidateLexMessageContentTypeInvalidType(t *testing.T) {
	invalidTypes := []string{
		"test",
		"JSON",
	}

	for _, invalidType := range invalidTypes {
		_, errs := validateLexMessageContentType(invalidType, "content_type")
		if len(errs) == 0 {
			t.Fatalf("%q should be an invalid Lex message content type, got %v", invalidType, errs)
		}
	}
}

func TestValidateLexMessageContentTypeValid(t *testing.T) {
	validTypes := []string{
		"PlainText",
		"SSML",
		"CustomPayload",
	}

	for _, validType := range validTypes {
		_, errs := validateLexMessageContentType(validType, "content_type")
		if len(errs) > 0 {
			t.Fatalf("%q should be a valid Lex message content type, got %v", validType, errs)
		}
	}
}
