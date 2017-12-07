# AWS SageMaker Example

This example takes the [example model provided by AWS](https://github.com/awslabs/amazon-sagemaker-examples/blob/master/advanced_functionality/scikit_bring_your_own/scikit_bring_your_own.ipynb)
to show how to deploy your own algorithm container with SageMaker using Terraform.


### Wrap model in Docker container and upload to [ECS](https://aws.amazon.com/ecs/)
```
    # get the SageMaker example model from AWS
    git clone https://github.com/awslabs/amazon-sagemaker-examples.git
    cd amazon-sagemaker-examples/advanced_functionality/scikit_bring_your_own/container/
    
    # export credentials for your account
    export AWS_ACCESS_KEY_ID=<your-access-key-id>
    export AWS_SECRET_ACCESS_KEY=<your-secret-access-key>
    
    # create docker container and push it to ECS
    ./build_and_push.sh foo
```

### Deploy model and run test prediction call

```
    # In the directory where this README is located, run the following
    terraform init
    terraform apply
    
    # make test call to the deployed model
    # go back to amazon-sagemaker-examples/advanced_functionality/scikit_bring_your_own/container/
    aws runtime.sagemaker invoke-endpoint --endpoint-name terraform-sagemaker-example \
        --body "`cat ./local_test/payload.csv`" --content-type "text/csv" "output.dat"
        
    # show the predicted values
    cat output.dat
    
    # destroy the deployed model
    terraform destroy
```
