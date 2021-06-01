# encoding=utf8

""" 注入Proto生成的类，提供额外的方法

    File Name: inject.py
    Description:

"""

from types import MethodType
from threading import local

InvocationLocal = local()

def inject(Invocation):
    """Inject
    """
    # Inject Invocation
    Invocation.current = InvocationCurrentProperty()
    Invocation.asCurrent = MethodType(asCurrent, Invocation)
    Invocation.getMetadataList = MethodType(getMetadataList, Invocation)

class InvocationCurrentProperty(object):
    """The invocation current property
    """
    def __get__(self, instance, cls):
        """Get current invocation
        """
        try:
            return InvocationLocal.__currentinvocation__
        except AttributeError:
            pass

def asCurrent(self):
    """Set this invocation as the current invocation
    """
    InvocationLocal.__currentinvocation__ = self

def getMetadataList(self):
    """Get the metadata list of the invocation (Could be used as the argument of metadata in gRPC client method)
    """
    return [ x.toTuple() for x in self.metadata ]
