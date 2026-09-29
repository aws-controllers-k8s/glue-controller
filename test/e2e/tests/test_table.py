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

"""Integration tests for the Glue Table.
"""

import logging
import time

import pytest
from acktest.k8s import resource as k8s
from acktest.k8s import condition
from acktest.resources import random_suffix_name
from e2e import CRD_GROUP, CRD_VERSION, load_glue_resource, service_marker
from e2e.replacement_values import REPLACEMENT_VALUES

from e2e.helper import GlueValidator

DATABASE_RESOURCE_PLURAL = 'databases'
RESOURCE_PLURAL = 'tables'

CREATE_WAIT_AFTER_SECONDS = 10
UPDATE_WAIT_AFTER_SECONDS = 10
DELETE_WAIT_AFTER_SECONDS = 10

@pytest.fixture(scope='module')
def simple_database(glue_client):
    database_name = random_suffix_name("table-db", 24)
    replacements = REPLACEMENT_VALUES.copy()
    replacements['DATABASE_NAME'] = database_name
    replacements['DATABASE_DESCRIPTION'] = "created by ACK e2e test"

    resource_data = load_glue_resource(
        'database',
        additional_replacements=replacements
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, DATABASE_RESOURCE_PLURAL,
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

@pytest.fixture(scope='module')
def simple_table(glue_client, simple_database):
    _, database_cr = simple_database
    database_name = database_cr['spec']['name']

    table_name = random_suffix_name("table", 24)
    replacements = REPLACEMENT_VALUES.copy()
    replacements['TABLE_NAME'] = table_name
    replacements['DATABASE_NAME'] = database_name
    replacements['TABLE_DESCRIPTION'] = "created by ACK e2e test"

    resource_data = load_glue_resource(
        'table',
        additional_replacements=replacements
    )
    logging.debug(resource_data)

    ref = k8s.CustomResourceReference(
        CRD_GROUP, CRD_VERSION, RESOURCE_PLURAL,
        table_name, namespace="default")
    k8s.create_custom_resource(ref, resource_data)

    time.sleep(CREATE_WAIT_AFTER_SECONDS)
    cr = k8s.wait_resource_consumed_by_controller(ref)

    assert cr is not None
    assert k8s.get_resource_exists(ref)

    yield(ref, cr, database_name)

    if k8s.get_resource_exists(ref):
        _, deleted = k8s.delete_custom_resource(
            ref,
            DELETE_WAIT_AFTER_SECONDS
        )
        assert deleted

@service_marker
@pytest.mark.canary
class TestTable():
    def test_create_update_delete_table(self, simple_table, glue_client):
        ref, cr, database_name = simple_table
        assert cr is not None
        assert 'spec' in cr
        assert 'name' in cr['spec']
        name = cr['spec']['name']

        condition.assert_synced(ref)

        cr = k8s.get_resource(ref)
        # catalogID is late initialized from the Glue API
        assert cr['spec']['catalogID'] is not None
        assert 'ackResourceMetadata' in cr['status']
        assert 'arn' in cr['status']['ackResourceMetadata']
        arn = cr['status']['ackResourceMetadata']['arn']
        assert arn.endswith(f":table/{database_name}/{name}")

        validator = GlueValidator(glue_client)

        latest = validator.get_table(database_name, name)
        assert latest is not None
        assert latest['Description'] == "created by ACK e2e test"
        assert latest['TableType'] == "EXTERNAL_TABLE"
        assert latest['Parameters']['classification'] == "csv"
        assert [k['Name'] for k in latest['PartitionKeys']] == ['dt']
        columns = latest['StorageDescriptor']['Columns']
        assert [(c['Name'], c['Type']) for c in columns] == [
            ('id', 'string'),
            ('value', 'bigint'),
        ]

        # The controller should not issue updates when nothing has changed
        version_id = latest['VersionId']
        time.sleep(UPDATE_WAIT_AFTER_SECONDS)
        latest = validator.get_table(database_name, name)
        assert latest['VersionId'] == version_id

        updates = {
            'spec': {
                'description': 'updated by ACK e2e test',
                'parameters': {
                    'classification': 'csv',
                    'compressionType': 'none',
                },
                'storageDescriptor': {
                    'columns': [
                        {'name': 'id', 'type': 'string', 'comment': 'unique identifier'},
                        {'name': 'value', 'type': 'bigint'},
                        {'name': 'label', 'type': 'string'},
                    ],
                },
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

        latest = validator.get_table(database_name, name)
        assert latest['Description'] == 'updated by ACK e2e test'
        assert latest['Parameters']['compressionType'] == 'none'
        columns = latest['StorageDescriptor']['Columns']
        assert [c['Name'] for c in columns] == ['id', 'value', 'label']

        _, deleted = k8s.delete_custom_resource(ref, DELETE_WAIT_AFTER_SECONDS)
        assert deleted
        assert not validator.table_exists(database_name, name)
