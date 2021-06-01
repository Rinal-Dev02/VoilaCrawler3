# encoding=utf8

""" The helpers

    File Name: helpers.py
    Description:

"""

from .spec import mKeyRequestStringify

def gRPCRequestStringify(f):
    """Set request stringify method
    Args:
        f(func (request, context)): A request stringify method
    """
    def decorator(method):
        """The decorator
        """
        setattr(method, mKeyRequestStringify, f)
        return method
    return decorator
