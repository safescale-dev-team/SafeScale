# coding: utf-8

from __future__ import absolute_import
from datetime import date, datetime  # noqa: F401

from typing import List, Dict  # noqa: F401

from rest.models.base_model_ import Model
from rest.models.reference import Reference
from rest.models.volume_attachment_response import VolumeAttachmentResponse
from rest.models.volume_speed import VolumeSpeed
from rest import util

from rest.models.reference import Reference  # noqa: E501
from rest.models.volume_attachment_response import VolumeAttachmentResponse  # noqa: E501
from rest.models.volume_speed import VolumeSpeed  # noqa: E501

class VolumeInspectResponse(Model):
    """NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).

    Do not edit the class manually.
    """

    def __init__(self, id=None, name=None, speed=None, size=None, host=None, mount_path=None, format=None, device=None, attachments=None):  # noqa: E501
        """VolumeInspectResponse - a model defined in OpenAPI

        :param id: The id of this VolumeInspectResponse.  # noqa: E501
        :type id: str
        :param name: The name of this VolumeInspectResponse.  # noqa: E501
        :type name: str
        :param speed: The speed of this VolumeInspectResponse.  # noqa: E501
        :type speed: VolumeSpeed
        :param size: The size of this VolumeInspectResponse.  # noqa: E501
        :type size: int
        :param host: The host of this VolumeInspectResponse.  # noqa: E501
        :type host: Reference
        :param mount_path: The mount_path of this VolumeInspectResponse.  # noqa: E501
        :type mount_path: str
        :param format: The format of this VolumeInspectResponse.  # noqa: E501
        :type format: str
        :param device: The device of this VolumeInspectResponse.  # noqa: E501
        :type device: str
        :param attachments: The attachments of this VolumeInspectResponse.  # noqa: E501
        :type attachments: List[VolumeAttachmentResponse]
        """
        self.openapi_types = {
            'id': str,
            'name': str,
            'speed': VolumeSpeed,
            'size': int,
            'host': Reference,
            'mount_path': str,
            'format': str,
            'device': str,
            'attachments': List[VolumeAttachmentResponse]
        }

        self.attribute_map = {
            'id': 'id',
            'name': 'name',
            'speed': 'speed',
            'size': 'size',
            'host': 'host',
            'mount_path': 'mountPath',
            'format': 'format',
            'device': 'device',
            'attachments': 'attachments'
        }

        self._id = id
        self._name = name
        self._speed = speed
        self._size = size
        self._host = host
        self._mount_path = mount_path
        self._format = format
        self._device = device
        self._attachments = attachments

    @classmethod
    def from_dict(cls, dikt) -> 'VolumeInspectResponse':
        """Returns the dict as a model

        :param dikt: A dict.
        :type: dict
        :return: The VolumeInspectResponse of this VolumeInspectResponse.  # noqa: E501
        :rtype: VolumeInspectResponse
        """
        return util.deserialize_model(dikt, cls)

    @property
    def id(self):
        """Gets the id of this VolumeInspectResponse.


        :return: The id of this VolumeInspectResponse.
        :rtype: str
        """
        return self._id

    @id.setter
    def id(self, id):
        """Sets the id of this VolumeInspectResponse.


        :param id: The id of this VolumeInspectResponse.
        :type id: str
        """

        self._id = id

    @property
    def name(self):
        """Gets the name of this VolumeInspectResponse.


        :return: The name of this VolumeInspectResponse.
        :rtype: str
        """
        return self._name

    @name.setter
    def name(self, name):
        """Sets the name of this VolumeInspectResponse.


        :param name: The name of this VolumeInspectResponse.
        :type name: str
        """

        self._name = name

    @property
    def speed(self):
        """Gets the speed of this VolumeInspectResponse.


        :return: The speed of this VolumeInspectResponse.
        :rtype: VolumeSpeed
        """
        return self._speed

    @speed.setter
    def speed(self, speed):
        """Sets the speed of this VolumeInspectResponse.


        :param speed: The speed of this VolumeInspectResponse.
        :type speed: VolumeSpeed
        """

        self._speed = speed

    @property
    def size(self):
        """Gets the size of this VolumeInspectResponse.


        :return: The size of this VolumeInspectResponse.
        :rtype: int
        """
        return self._size

    @size.setter
    def size(self, size):
        """Sets the size of this VolumeInspectResponse.


        :param size: The size of this VolumeInspectResponse.
        :type size: int
        """

        self._size = size

    @property
    def host(self):
        """Gets the host of this VolumeInspectResponse.


        :return: The host of this VolumeInspectResponse.
        :rtype: Reference
        """
        return self._host

    @host.setter
    def host(self, host):
        """Sets the host of this VolumeInspectResponse.


        :param host: The host of this VolumeInspectResponse.
        :type host: Reference
        """

        self._host = host

    @property
    def mount_path(self):
        """Gets the mount_path of this VolumeInspectResponse.


        :return: The mount_path of this VolumeInspectResponse.
        :rtype: str
        """
        return self._mount_path

    @mount_path.setter
    def mount_path(self, mount_path):
        """Sets the mount_path of this VolumeInspectResponse.


        :param mount_path: The mount_path of this VolumeInspectResponse.
        :type mount_path: str
        """

        self._mount_path = mount_path

    @property
    def format(self):
        """Gets the format of this VolumeInspectResponse.


        :return: The format of this VolumeInspectResponse.
        :rtype: str
        """
        return self._format

    @format.setter
    def format(self, format):
        """Sets the format of this VolumeInspectResponse.


        :param format: The format of this VolumeInspectResponse.
        :type format: str
        """

        self._format = format

    @property
    def device(self):
        """Gets the device of this VolumeInspectResponse.


        :return: The device of this VolumeInspectResponse.
        :rtype: str
        """
        return self._device

    @device.setter
    def device(self, device):
        """Sets the device of this VolumeInspectResponse.


        :param device: The device of this VolumeInspectResponse.
        :type device: str
        """

        self._device = device

    @property
    def attachments(self):
        """Gets the attachments of this VolumeInspectResponse.


        :return: The attachments of this VolumeInspectResponse.
        :rtype: List[VolumeAttachmentResponse]
        """
        return self._attachments

    @attachments.setter
    def attachments(self, attachments):
        """Sets the attachments of this VolumeInspectResponse.


        :param attachments: The attachments of this VolumeInspectResponse.
        :type attachments: List[VolumeAttachmentResponse]
        """

        self._attachments = attachments
