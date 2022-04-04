#!/usr/bin/env python3

import argparse
import random
import boto3

def main(argv):
    """
    Simple Python script to populate an S3 bucket with a number of objects.
    The S3 bucket must already exist and have versioning enabled.
    AWS credentials must be set in the environment.
    """
    p = argparse.ArgumentParser(description="script to populate an existing S3 Bucket with objects")
    p.add_argument("bucket")
    p.add_argument("-n", "--number-of-objects", action="store", default=100)

    args = p.parse_args(args=argv)

    populate_bucket(args.bucket, int(args.number_of_objects))

def populate_bucket(bucket, num_objects):
    client = boto3.client('s3')

    print("populating %s with %d objects..." % (bucket, num_objects))

    for i in range(num_objects):
        key = "object-%d" % (i)

        # Each object has between 50 and 100 versions.
        for j in range(random.randint(50, 100)):
            contents = "data.%d" % (j)
            client.put_object(Bucket=bucket, Key=key, Body=contents.encode('utf-8'))

            # 10% chance of the object then being deleted.
            if random.random() < 0.10:
                client.delete_object(Bucket=bucket, Key=key)

    print("done!")

if __name__ == '__main__':
    import sys
    main(sys.argv[1:])