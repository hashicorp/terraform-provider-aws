# Unit Tests

Unlike acceptance tests, unit tests do not access AWS and are focused on a function (or method). Because of this, they are quick and cheap to run.

In designing a resource's implementation, isolate complex bits from AWS bits so that they can be tested through a unit test. We encourage more unit tests in the provider.

## In context

To help place unit testing in context, here is an overview of the Terraform AWS Provider's three types of tests.

1. [**Acceptance tests**](running-and-writing-acceptance-tests.md) are end-to-end evaluations of interactions with AWS. They validate functionalities like creating, reading, and destroying resources within AWS.
2. **Unit tests** (_You are here!_) focus on testing isolated units of code within the software, typically at the function level. They assess functionalities solely within the provider itself.
3. [**Continuous integration tests**](continuous-integration.md) encompass a suite of automated tests that are executed on every pull request and include linting, compiling code, running unit tests, and performing static analysis.

## What to test

Utilitarian functions that carry out processing within the provider, that don't need contact with AWS, are candidates for unit testing. In specific, any moderately to very complex utilitarian function should have a corresponding unit test.

Rather than being a burden, using unit tests often saves time (and money) during development because they can be run quickly and locally. In addition, rather than mentally processing all edge cases, you can use test cases to refine the function's behavior.

Cut and dry functions using well-used patterns, like typical flatteners and expanders (flex functions), don't need unit testing. However, if the flex functions are complex or intricate, they should be unit tested.

## Where they go

Unit tests can be placed in a test file with acceptance tests if they mainly relate to a single resource. However, if it makes sense to do so, functions and their unit tests can be moved to their files. Reasons you might split them from acceptance test files include length of the acceptance test files, or several resources using them.

Where unit tests are included with acceptance tests in resource test files, they should be placed at the top of the file, before the first acceptance test.

## Example

This is an example of a unit test.

Its name is prefixed with "Test" but not "TestAcc" like an acceptance test.

```go
func TestExampleUnitTest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		TestName string
		Input    string
		Expected string
		Error    bool
	}{
		{
			TestName: "empty",
			Input:    "",
			Expected: "",
			Error:    true,
		},
		{
			TestName: "descriptive name",
			Input:    "some input",
			Expected: "some output",
			Error:    false,
		},
		{
			TestName: "another descriptive name",
			Input:    "more input",
			Expected: "more output",
			Error:    false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()
			got, err := tfrds.FunctionFromResource(testCase.Input)

			if err != nil && !testCase.Error {
				t.Errorf("got error (%s), expected no error", err)
			}

			if err == nil && testCase.Error {
				t.Errorf("got (%s) and no error, expected error", got)
			}

			if got != testCase.Expected {
				t.Errorf("got %s, expected %s", got, testCase.Expected)
			}
		})
	}
}
```
