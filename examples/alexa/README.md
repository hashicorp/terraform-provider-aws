# Alexa Example

This example will show you how to use Terraform to enable the Alexa permission on a lambda function.

This example codifies [this guide](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/developing-an-alexa-skill-as-a-lambda-function), which does the following:

- Creates an IAM role and role policy for your Lambda which enables logging.
- Creates the Amazon Skills Kit `alexa-skills-kit-color-expert-python` lambda
- Adds the Alexa permission to the lambda.

Once the lambda is created with Terraform you can head over to the [Amazon Developer Portal](https://developer.amazon.com) to register an Alexa Skill which will use the provided [Sample Interaction Model for the Color Expert Blueprint](https://developer.amazon.com/public/solutions/alexa/alexa-skills-kit/docs/developing-an-alexa-skill-as-a-lambda-function#sample-interaction-model-for-the-color-expert-blueprint) and the ARN of the lambda (which is a Terraform output of this example).

## Run the example

From inside of this directory:

```bash
export AWS_ACCESS_KEY_ID=<this is a secret>
export AWS_SECRET_ACCESS_KEY=<this is a secret>
terraform init
terraform plan -out theplan
terraform apply theplan
```

## Remove the example

```bash
terraform destroy
```

Go to the console and remove the CloudWatch Log Group `/aws/lambda/terraform_lambda_alexa_example` which is created by AWS.
