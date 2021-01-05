# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.cluster_response import ClusterResponse
from rest import util

from rest.models.cluster_response import ClusterResponse  # noqa: E501

class ClusterListResponse(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, clusters=None):  # noqa: E501
        """ClusterListResponse - a model defined in OpenAPI

        :param clusters: The clusters of this ClusterListResponse.  # noqa: E501
        :type clusters: List[ClusterResponse]
        """
        self.openapi_types = {
            'clusters': List[ClusterResponse]
        }

        self.attribute_map = {
            'clusters': 'clusters'
        }

        self._clusters = clusters

    @classmethod
    def from_dict(cls, dikt) -> 'ClusterListResponse':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The ClusterListResponse of this ClusterListResponse.  # noqa: E501
        :rtype: ClusterListResponse
        """
        return util.deserialize_model(dikt, cls)

    @property
    def clusters(self):
        """Gets the clusters of this ClusterListResponse.


        :return: The clusters of this ClusterListResponse.
        :rtype: List[ClusterResponse]
        """
        return self._clusters

    @clusters.setter
    def clusters(self, clusters):
        """Sets the clusters of this ClusterListResponse.


        :param clusters: The clusters of this ClusterListResponse.
        :type clusters: List[ClusterResponse]
        """

        self._clusters = clusters
