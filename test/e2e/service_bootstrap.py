# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
"""Bootstraps the resources required to run the Glue integration tests.
"""
import os
import boto3
import logging
from zipfile import ZipFile

from botocore.exceptions import ClientError
from acktest.bootstrapping import Resources, BootstrapFailureException
from acktest.bootstrapping.iam import Role
from acktest.bootstrapping.s3 import Bucket

from e2e import bootstrap_directory
from e2e.bootstrap_resources import BootstrapResources

GLUE_JOB_FILE = "main.py"
GLUE_JOB_FILE_PATH = f"./resources/glue_job/{GLUE_JOB_FILE}"

def zip_script_file(src: str, dst: str):
    with ZipFile(dst, 'w') as zipf:
        zipf.write(src, arcname=src)

def upload_script_to_bucket(file_path: str, bucket_name: str):
    object_name = os.path.basename(file_path)

    s3_client = boto3.client('s3')
    try:
        s3_client.upload_file(
            file_path,
            bucket_name,
            object_name,
        )
    except ClientError as e:
        logging.error(e)

    logging.info(f"Uploaded {file_path} to bucket {bucket_name}")

def service_bootstrap() -> Resources:
    logging.getLogger().setLevel(logging.INFO)

    resources = BootstrapResources(
        # TODO: Add bootstrapping when you have defined the resources
        JobBucket=Bucket(
            "ack-glue-controller-tests",
        ),
        JobRole=Role(
            "ack-glue-controller-basic-role",
            principal_service="glue.amazonaws.com",
            managed_policies=["arn:aws:iam::aws:policy/service-role/AWSGlueServiceRole"],
        ),
    )

    try:
        resources.bootstrap()
        upload_script_to_bucket(
            GLUE_JOB_FILE_PATH,
            resources.JobBucket.name,
        )
    except BootstrapFailureException as ex:
        exit(254)

    return resources

if __name__ == "__main__":
    config = service_bootstrap()
    # Write config to current directory by default
    config.serialize(bootstrap_directory)
