# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.security_group_rule import SecurityGroupRule
from rest import util

from rest.models.security_group_rule import SecurityGroupRule  # noqa: E501

class SecurityGroupResponse(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, id=None, name=None, description=None, rules=None):  # noqa: E501
        """SecurityGroupResponse - a model defined in OpenAPI

        :param id: The id of this SecurityGroupResponse.  # noqa: E501
        :type id: str
        :param name: The name of this SecurityGroupResponse.  # noqa: E501
        :type name: str
        :param description: The description of this SecurityGroupResponse.  # noqa: E501
        :type description: str
        :param rules: The rules of this SecurityGroupResponse.  # noqa: E501
        :type rules: List[SecurityGroupRule]
        """
        self.openapi_types = {
            'id': str,
            'name': str,
            'description': str,
            'rules': List[SecurityGroupRule]
        }

        self.attribute_map = {
            'id': 'id',
            'name': 'name',
            'description': 'description',
            'rules': 'rules'
        }

        self._id = id
        self._name = name
        self._description = description
        self._rules = rules

    @classmethod
    def from_dict(cls, dikt) -> 'SecurityGroupResponse':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The SecurityGroupResponse of this SecurityGroupResponse.  # noqa: E501
        :rtype: SecurityGroupResponse
        """
        return util.deserialize_model(dikt, cls)

    @property
    def id(self):
        """Gets the id of this SecurityGroupResponse.


        :return: The id of this SecurityGroupResponse.
        :rtype: str
        """
        return self._id

    @id.setter
    def id(self, id):
        """Sets the id of this SecurityGroupResponse.


        :param id: The id of this SecurityGroupResponse.
        :type id: str
        """

        self._id = id

    @property
    def name(self):
        """Gets the name of this SecurityGroupResponse.


        :return: The name of this SecurityGroupResponse.
        :rtype: str
        """
        return self._name

    @name.setter
    def name(self, name):
        """Sets the name of this SecurityGroupResponse.


        :param name: The name of this SecurityGroupResponse.
        :type name: str
        """

        self._name = name

    @property
    def description(self):
        """Gets the description of this SecurityGroupResponse.


        :return: The description of this SecurityGroupResponse.
        :rtype: str
        """
        return self._description

    @description.setter
    def description(self, description):
        """Sets the description of this SecurityGroupResponse.


        :param description: The description of this SecurityGroupResponse.
        :type description: str
        """

        self._description = description

    @property
    def rules(self):
        """Gets the rules of this SecurityGroupResponse.


        :return: The rules of this SecurityGroupResponse.
        :rtype: List[SecurityGroupRule]
        """
        return self._rules

    @rules.setter
    def rules(self, rules):
        """Sets the rules of this SecurityGroupResponse.


        :param rules: The rules of this SecurityGroupResponse.
        :type rules: List[SecurityGroupRule]
        """

        self._rules = rules
