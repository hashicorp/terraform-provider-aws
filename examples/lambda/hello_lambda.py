import os

def lambda_handler(event, context):
    # This will show up in CloudWatch
    print("Value of 'foo': " + os.environ['foo'])
    return 'Hello from Lambda!'
