# Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
# 	 http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.

"""Integration tests for the Glue Database.
"""

import logging
import time

import pytest
from acktest import tags
from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import CRD_GROUP, CRD_VERSION, load_glue_resource, service_marker
from e2e.replacement_values import REPLACEMENT_VALUES

from e2e.helper import GlueValidator

RESOURCE_PLURAL = 'databases'

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope='module')
def simple_database(glue_client):
    database_name = random_suffix_name("database", 24)
    replacements = REPLACEMENT_VALUES.copy()
    replacements['DATABASE_NAME'] = database_name
    replacements['DATABASE_DESCRIPTION'] = "created by ACK e2e test"

    resource_data = load_glue_resource(
        'database',
        additional_replacements=replacements
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        database_name, namespace="default")
    k8s.create_custom_resource(ref, resource_data)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield(ref, cr)

    if k8s.get_resource_exists(ref):
        _, deleted = k8s.delete_custom_resource(
            ref,
            DELETE_WAIT_AFTER_SECONDS
        )
        assert deleted

@service_marker
@pytest.mark.canary
class TestDatabase():
    def test_create_delete_simple_database(self, simple_database, glue_client):
        ref, cr = simple_database
        assert cr is not None
        assert 'spec' in cr
        assert 'name' in cr['spec']
        name = cr['spec']['name']
        assert 'status' in cr
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']

        validator = GlueValidator(glue_client)

        latest = validator.get_database(name)
        assert latest is not None
        assert latest['Description'] == "created by ACK e2e test"

        condition.assert_synced(ref)

        cr = k8s.get_resource(ref)
        assert cr['spec']['catalogID'] is not None

        desired_tags = cr['spec']['tags']
        latest_tags = validator.database_list_tags(arn)
        tags.assert_ack_system_tags(
            tags=latest_tags,
        )
        tags.assert_equal_without_ack_tags(
            expected=desired_tags,
            actual=latest_tags,
        )

        updates = {
            'spec': {
                'description': 'updated by ACK e2e test',
                'tags': {
                    'newKey': 'newVal',
                }
            }
        }
        k8s.patch_custom_resource(ref, updates)
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)
        assert k8s.wait_on_condition(
            ref,
            "ACK.ResourceSynced",
            "True",
            wait_periods=UPDATE_WAIT_AFTER_SECONDS,
        )

        cr = k8s.get_resource(ref)

        latest = validator.get_database(name)
        assert latest['Description'] == 'updated by ACK e2e test'

        latest_tags = validator.database_list_tags(arn)
        desired_tags = cr['spec']['tags']
        tags.assert_ack_system_tags(
            tags=latest_tags,
        )
        tags.assert_equal_without_ack_tags(
            expected=desired_tags,
            actual=latest_tags,
        )
