# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.key_value import KeyValue
from rest import util

from rest.models.key_value import KeyValue  # noqa: E501

class TenantIdentity(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, user=None, app_key=None, domain=None):  # noqa: E501
        """TenantIdentity - a model defined in OpenAPI

        :param user: The user of this TenantIdentity.  # noqa: E501
        :type user: KeyValue
        :param app_key: The app_key of this TenantIdentity.  # noqa: E501
        :type app_key: KeyValue
        :param domain: The domain of this TenantIdentity.  # noqa: E501
        :type domain: KeyValue
        """
        self.openapi_types = {
            'user': KeyValue,
            'app_key': KeyValue,
            'domain': KeyValue
        }

        self.attribute_map = {
            'user': 'user',
            'app_key': 'appKey',
            'domain': 'domain'
        }

        self._user = user
        self._app_key = app_key
        self._domain = domain

    @classmethod
    def from_dict(cls, dikt) -> 'TenantIdentity':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The TenantIdentity of this TenantIdentity.  # noqa: E501
        :rtype: TenantIdentity
        """
        return util.deserialize_model(dikt, cls)

    @property
    def user(self):
        """Gets the user of this TenantIdentity.


        :return: The user of this TenantIdentity.
        :rtype: KeyValue
        """
        return self._user

    @user.setter
    def user(self, user):
        """Sets the user of this TenantIdentity.


        :param user: The user of this TenantIdentity.
        :type user: KeyValue
        """

        self._user = user

    @property
    def app_key(self):
        """Gets the app_key of this TenantIdentity.


        :return: The app_key of this TenantIdentity.
        :rtype: KeyValue
        """
        return self._app_key

    @app_key.setter
    def app_key(self, app_key):
        """Sets the app_key of this TenantIdentity.


        :param app_key: The app_key of this TenantIdentity.
        :type app_key: KeyValue
        """

        self._app_key = app_key

    @property
    def domain(self):
        """Gets the domain of this TenantIdentity.


        :return: The domain of this TenantIdentity.
        :rtype: KeyValue
        """
        return self._domain

    @domain.setter
    def domain(self, domain):
        """Sets the domain of this TenantIdentity.


        :param domain: The domain of this TenantIdentity.
        :type domain: KeyValue
        """

        self._domain = domain
