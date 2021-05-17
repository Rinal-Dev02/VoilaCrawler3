# encoding=utf8

import logging

from google.protobuf import symbol_database
from protobuf.annotations_pb2 import auth
from protobuf.options.auth_pb2 import NoAuth

_symdb = symbol_database.Default()

class ServiceDesc(object):
    """The service description
    """
    def __init__(self, descriptor, package):
        """Create a new ServiceDesc
        """
        self.package = package
        self.descriptor = descriptor
        self.methods = {
            methodDescriptor.name: MethodDesc(methodDescriptor, descriptor.name, package) for methodDescriptor in descriptor.method
            }

    @property
    def name(self):
        """Get the service name
        """
        return self.descriptor.name

class MethodDesc(object):
    """The method description
    """
    def __init__(self, descriptor, service, package):
        """Create a new MethodDesc
        """
        self.logger = logging.getLogger("%s.%s:%s" % (package, service, descriptor.name))
        self.package = package
        self.service = service
        self.descriptor = descriptor
        # Get auth rule (NoAuth == None)
        self.authRule = MethodDesc.getAuthRule(descriptor)

    @property
    def name(self):
        """Get the method name
        """
        return self.descriptor.name

    @property
    def fullname(self):
        """Get the method full name
        """
        return "/%s.%s/%s" % (self.package, self.service, self.name)

    @property
    def isServerStreamming(self):
        """Check if is server streamming (The return value is a generator)
        """
        return self.descriptor.server_streaming

    def response(self):
        """Return a new empty response object
        """
        try:
            if self.descriptor.server_streaming:
                # Empty generator
                return (_ for _ in ())
            else:
                # Empty object
                return _symdb.GetSymbol(self.descriptor.output_type[1: ])()
        except KeyError:
            raise ValueError("Response type [%s] not found" % self.descriptor.output_type)

    @classmethod
    def getAuthRule(cls, descriptor):
        """Get the auth rule defined on this method.
        No auth rule defined or level == NoAuth are both treat as no auth rule is defined.

        Args:
            descriptor(google.protobuf.descriptor_pb2.MethodDescriptorProto): The method descriptor
        Returns:
            protobuf.options.auth_pb2.AuthRule
        """
        if descriptor.options.HasExtension(auth):
            authRule = descriptor.options.Extensions[auth]
            if authRule.level != NoAuth:
                return authRule
