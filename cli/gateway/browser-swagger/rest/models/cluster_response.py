# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.cluster_composite import ClusterComposite
from rest.models.cluster_controlplane import ClusterControlplane
from rest.models.cluster_defaults import ClusterDefaults
from rest.models.cluster_identity import ClusterIdentity
from rest.models.cluster_network import ClusterNetwork
from rest.models.cluster_state import ClusterState
from rest.models.feature_list_response import FeatureListResponse
from rest.models.host import Host
from rest import util

from rest.models.cluster_composite import ClusterComposite  # noqa: E501
from rest.models.cluster_controlplane import ClusterControlplane  # noqa: E501
from rest.models.cluster_defaults import ClusterDefaults  # noqa: E501
from rest.models.cluster_identity import ClusterIdentity  # noqa: E501
from rest.models.cluster_network import ClusterNetwork  # noqa: E501
from rest.models.cluster_state import ClusterState  # noqa: E501
from rest.models.feature_list_response import FeatureListResponse  # noqa: E501
from rest.models.host import Host  # noqa: E501

class ClusterResponse(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, identity=None, network=None, masters=None, nodes=None, disabled_features=None, installed_features=None, defaults=None, state=None, composite=None, controlplane=None):  # noqa: E501
        """ClusterResponse - a model defined in OpenAPI

        :param identity: The identity of this ClusterResponse.  # noqa: E501
        :type identity: ClusterIdentity
        :param network: The network of this ClusterResponse.  # noqa: E501
        :type network: ClusterNetwork
        :param masters: The masters of this ClusterResponse.  # noqa: E501
        :type masters: List[Host]
        :param nodes: The nodes of this ClusterResponse.  # noqa: E501
        :type nodes: List[Host]
        :param disabled_features: The disabled_features of this ClusterResponse.  # noqa: E501
        :type disabled_features: FeatureListResponse
        :param installed_features: The installed_features of this ClusterResponse.  # noqa: E501
        :type installed_features: FeatureListResponse
        :param defaults: The defaults of this ClusterResponse.  # noqa: E501
        :type defaults: ClusterDefaults
        :param state: The state of this ClusterResponse.  # noqa: E501
        :type state: ClusterState
        :param composite: The composite of this ClusterResponse.  # noqa: E501
        :type composite: ClusterComposite
        :param controlplane: The controlplane of this ClusterResponse.  # noqa: E501
        :type controlplane: ClusterControlplane
        """
        self.openapi_types = {
            'identity': ClusterIdentity,
            'network': ClusterNetwork,
            'masters': List[Host],
            'nodes': List[Host],
            'disabled_features': FeatureListResponse,
            'installed_features': FeatureListResponse,
            'defaults': ClusterDefaults,
            'state': ClusterState,
            'composite': ClusterComposite,
            'controlplane': ClusterControlplane
        }

        self.attribute_map = {
            'identity': 'identity',
            'network': 'network',
            'masters': 'masters',
            'nodes': 'nodes',
            'disabled_features': 'disabledFeatures',
            'installed_features': 'installedFeatures',
            'defaults': 'defaults',
            'state': 'state',
            'composite': 'composite',
            'controlplane': 'controlplane'
        }

        self._identity = identity
        self._network = network
        self._masters = masters
        self._nodes = nodes
        self._disabled_features = disabled_features
        self._installed_features = installed_features
        self._defaults = defaults
        self._state = state
        self._composite = composite
        self._controlplane = controlplane

    @classmethod
    def from_dict(cls, dikt) -> 'ClusterResponse':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The ClusterResponse of this ClusterResponse.  # noqa: E501
        :rtype: ClusterResponse
        """
        return util.deserialize_model(dikt, cls)

    @property
    def identity(self):
        """Gets the identity of this ClusterResponse.


        :return: The identity of this ClusterResponse.
        :rtype: ClusterIdentity
        """
        return self._identity

    @identity.setter
    def identity(self, identity):
        """Sets the identity of this ClusterResponse.


        :param identity: The identity of this ClusterResponse.
        :type identity: ClusterIdentity
        """

        self._identity = identity

    @property
    def network(self):
        """Gets the network of this ClusterResponse.


        :return: The network of this ClusterResponse.
        :rtype: ClusterNetwork
        """
        return self._network

    @network.setter
    def network(self, network):
        """Sets the network of this ClusterResponse.


        :param network: The network of this ClusterResponse.
        :type network: ClusterNetwork
        """

        self._network = network

    @property
    def masters(self):
        """Gets the masters of this ClusterResponse.


        :return: The masters of this ClusterResponse.
        :rtype: List[Host]
        """
        return self._masters

    @masters.setter
    def masters(self, masters):
        """Sets the masters of this ClusterResponse.


        :param masters: The masters of this ClusterResponse.
        :type masters: List[Host]
        """

        self._masters = masters

    @property
    def nodes(self):
        """Gets the nodes of this ClusterResponse.


        :return: The nodes of this ClusterResponse.
        :rtype: List[Host]
        """
        return self._nodes

    @nodes.setter
    def nodes(self, nodes):
        """Sets the nodes of this ClusterResponse.


        :param nodes: The nodes of this ClusterResponse.
        :type nodes: List[Host]
        """

        self._nodes = nodes

    @property
    def disabled_features(self):
        """Gets the disabled_features of this ClusterResponse.


        :return: The disabled_features of this ClusterResponse.
        :rtype: FeatureListResponse
        """
        return self._disabled_features

    @disabled_features.setter
    def disabled_features(self, disabled_features):
        """Sets the disabled_features of this ClusterResponse.


        :param disabled_features: The disabled_features of this ClusterResponse.
        :type disabled_features: FeatureListResponse
        """

        self._disabled_features = disabled_features

    @property
    def installed_features(self):
        """Gets the installed_features of this ClusterResponse.


        :return: The installed_features of this ClusterResponse.
        :rtype: FeatureListResponse
        """
        return self._installed_features

    @installed_features.setter
    def installed_features(self, installed_features):
        """Sets the installed_features of this ClusterResponse.


        :param installed_features: The installed_features of this ClusterResponse.
        :type installed_features: FeatureListResponse
        """

        self._installed_features = installed_features

    @property
    def defaults(self):
        """Gets the defaults of this ClusterResponse.


        :return: The defaults of this ClusterResponse.
        :rtype: ClusterDefaults
        """
        return self._defaults

    @defaults.setter
    def defaults(self, defaults):
        """Sets the defaults of this ClusterResponse.


        :param defaults: The defaults of this ClusterResponse.
        :type defaults: ClusterDefaults
        """

        self._defaults = defaults

    @property
    def state(self):
        """Gets the state of this ClusterResponse.


        :return: The state of this ClusterResponse.
        :rtype: ClusterState
        """
        return self._state

    @state.setter
    def state(self, state):
        """Sets the state of this ClusterResponse.


        :param state: The state of this ClusterResponse.
        :type state: ClusterState
        """

        self._state = state

    @property
    def composite(self):
        """Gets the composite of this ClusterResponse.


        :return: The composite of this ClusterResponse.
        :rtype: ClusterComposite
        """
        return self._composite

    @composite.setter
    def composite(self, composite):
        """Sets the composite of this ClusterResponse.


        :param composite: The composite of this ClusterResponse.
        :type composite: ClusterComposite
        """

        self._composite = composite

    @property
    def controlplane(self):
        """Gets the controlplane of this ClusterResponse.


        :return: The controlplane of this ClusterResponse.
        :rtype: ClusterControlplane
        """
        return self._controlplane

    @controlplane.setter
    def controlplane(self, controlplane):
        """Sets the controlplane of this ClusterResponse.


        :param controlplane: The controlplane of this ClusterResponse.
        :type controlplane: ClusterControlplane
        """

        self._controlplane = controlplane
