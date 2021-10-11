# encoding=utf8

""" The framework

    File Name: __init__.py
    Description:

"""

from .desc import ServiceDesc, MethodDesc
from .helpers import gRPCRequestStringify
from .invocation_pb2 import Invocation, InvokeTarget, InvokeAuthData, InvokeMetadata
from .serializer import Serializer, OptionsTimestamp2Datetime, OptionsDuration2Timedelta, OptionsDuration2FloatSeconds
from .interceptor import EmptyServerInterceptor, EmptyClientInterceptor
from .deserializer import Deserializer
# from .data_pb2 import MaxInvokeRate
# from .service_pb2_grpc import RPCStub

# Inject
from .inject import inject
inject(Invocation)

__all__ = [
    "ServiceDesc", "MethodDesc",
    "gRPCRequestStringify",
    # "MaxInvokeRate",
    "Invocation", "InvokeTarget", "InvokeAuthData", "InvokeMetadata",
    "Serializer", "OptionsTimestamp2Datetime", "OptionsDuration2Timedelta", "OptionsDuration2FloatSeconds",
    "EmptyServerInterceptor", "EmptyClientInterceptor",
    "Deserializer",
    # "RPCStub",
]
