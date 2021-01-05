import connexion
import six

from rest.models.rpc_status import RpcStatus  # noqa: E501
from rest.models.tenant_list import TenantList  # noqa: E501
from rest.models.tenant_name import TenantName  # noqa: E501
from rest import util


def tenant_service_get():  # noqa: E501
    """tenant_service_get

     # noqa: E501


    :rtype: TenantName
    """
    return 'do some magic!'


def tenant_service_list():  # noqa: E501
    """tenant_service_list

     # noqa: E501


    :rtype: TenantList
    """
    return 'do some magic!'


def tenant_service_set():  # noqa: E501
    """tenant_service_set

     # noqa: E501


    :rtype: object
    """
    return 'do some magic!'
