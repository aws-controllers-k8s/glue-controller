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

"""Helper functions for Glue e2e tests
"""

class GlueValidator:
    def __init__(self, glue_client):
        self.glue_client = glue_client
    
    def get_job(self, job_name):
        try:
            response = self.glue_client.get_job(JobName=job_name)
            return response['Job']
        except self.glue_client.exceptions.EntityNotFoundException:
            return None
        
    def job_exists(self, user_pool_id):
        response = self.get_job(user_pool_id)
        return response is not None

    def job_list_tags(self, job_arn):
        try:
            response = self.glue_client.get_tags(ResourceArn=job_arn)
            return response['Tags']
        except self.glue_client.exceptions.EntityNotFoundException:
            return None