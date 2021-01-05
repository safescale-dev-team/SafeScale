# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.tenant_object_storage import TenantObjectStorage
from rest import util

from rest.models.tenant_object_storage import TenantObjectStorage  # noqa: E501

class TenantMetadata(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, storage=None, bucket_name=None, crypt=None):  # noqa: E501
        """TenantMetadata - a model defined in OpenAPI

        :param storage: The storage of this TenantMetadata.  # noqa: E501
        :type storage: TenantObjectStorage
        :param bucket_name: The bucket_name of this TenantMetadata.  # noqa: E501
        :type bucket_name: str
        :param crypt: The crypt of this TenantMetadata.  # noqa: E501
        :type crypt: bool
        """
        self.openapi_types = {
            'storage': TenantObjectStorage,
            'bucket_name': str,
            'crypt': bool
        }

        self.attribute_map = {
            'storage': 'storage',
            'bucket_name': 'bucketName',
            'crypt': 'crypt'
        }

        self._storage = storage
        self._bucket_name = bucket_name
        self._crypt = crypt

    @classmethod
    def from_dict(cls, dikt) -> 'TenantMetadata':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The TenantMetadata of this TenantMetadata.  # noqa: E501
        :rtype: TenantMetadata
        """
        return util.deserialize_model(dikt, cls)

    @property
    def storage(self):
        """Gets the storage of this TenantMetadata.


        :return: The storage of this TenantMetadata.
        :rtype: TenantObjectStorage
        """
        return self._storage

    @storage.setter
    def storage(self, storage):
        """Sets the storage of this TenantMetadata.


        :param storage: The storage of this TenantMetadata.
        :type storage: TenantObjectStorage
        """

        self._storage = storage

    @property
    def bucket_name(self):
        """Gets the bucket_name of this TenantMetadata.


        :return: The bucket_name of this TenantMetadata.
        :rtype: str
        """
        return self._bucket_name

    @bucket_name.setter
    def bucket_name(self, bucket_name):
        """Sets the bucket_name of this TenantMetadata.


        :param bucket_name: The bucket_name of this TenantMetadata.
        :type bucket_name: str
        """

        self._bucket_name = bucket_name

    @property
    def crypt(self):
        """Gets the crypt of this TenantMetadata.


        :return: The crypt of this TenantMetadata.
        :rtype: bool
        """
        return self._crypt

    @crypt.setter
    def crypt(self, crypt):
        """Sets the crypt of this TenantMetadata.


        :param crypt: The crypt of this TenantMetadata.
        :type crypt: bool
        """

        self._crypt = crypt
