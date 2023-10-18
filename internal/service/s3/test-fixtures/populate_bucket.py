#!/usr/bin/env python3
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0


import argparse
import random
import boto3

from datetime import date, datetime, timedelta

def main(argv):
    """
    Simple Python script to populate an S3 bucket with a number of objects.
    The S3 bucket must already exist and have versioning enabled.
    AWS credentials must be set in the environment.
    """
    p = argparse.ArgumentParser(description="script to populate an existing S3 Bucket with objects")
    p.add_argument("bucket")
    p.add_argument("-n", "--number-of-objects", action="store", default=100)
    p.add_argument("-l", "--object-lock-enabled", action="store_true", default=False)

    args = p.parse_args(args=argv)

    populate_bucket(args.bucket, int(args.number_of_objects), bool(args.object_lock_enabled))

def populate_bucket(bucket, num_objects, obj_lock):
    client = boto3.client('s3')

    print("populating %s with %d objects..." % (bucket, num_objects))

    in_10_days = date.today() + timedelta(days=10)
    retail_until = datetime(in_10_days.year, in_10_days.month, in_10_days.day)

    for i in range(num_objects):
        key = "object-%d" % (i)

        # Each object has between 50 and 100 versions.
        for j in range(random.randint(50, 100)):
            contents = "data.%d" % (j)
            args = {
                "Bucket": bucket,
                "Key": key,
                "Body": contents.encode('utf-8'),
            }

            if obj_lock:
                # 5% chance of the object being locked in Governance mode.
                chance = random.random()
                if chance < 0.05:
                    args["ObjectLockMode"] = "GOVERNANCE"
                    args["ObjectLockRetainUntilDate"] = retail_until
                # Or a 5% chance of having a legal hold in place.
                elif chance < 0.10:
                    args["ObjectLockLegalHoldStatus"] = "ON"

            client.put_object(**args)

            # 10% chance of the object then being deleted.
            if random.random() < 0.10:
                client.delete_object(Bucket=bucket, Key=key)

    print("done!")

if __name__ == '__main__':
    import sys
    main(sys.argv[1:])