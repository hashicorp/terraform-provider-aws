package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
)

func TestValidateStringMinMaxRegexTooShort(t *testing.T) {
	_, errs := validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex)("", "name")
	if len(errs) == 0 {
		t.Fatalf("an empty string should return an error")
	}
}

func TestValidateStringMinMaxRegexTooLong(t *testing.T) {
	_, errs := validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex)(acctest.RandString(lexNameMaxLength+1), "name")
	if len(errs) == 0 {
		t.Fatalf("a %d character string should return an error", lexNameMaxLength+1)
	}
}

func TestValidateStringMinMaxRegexInvalidCharacters(t *testing.T) {
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
		_, errs := validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex)(invalidName, "name")
		if len(errs) == 0 {
			t.Fatalf("%q should be an invalid Lex name", invalidName)
		}
	}
}

func TestValidateStringMinMaxRegexValid(t *testing.T) {
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
		_, errs := validateStringMinMaxRegex(lexNameMinLength, lexNameMaxLength, lexNameRegex)(validName, "name")
		if len(errs) > 0 {
			t.Fatalf("%q should be a valid Lex name, got %v", validName, errs)
		}
	}
}
