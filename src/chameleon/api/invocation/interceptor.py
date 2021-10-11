# encoding=utf8
# pylint: disable=W0703

""" The python gRPC interceptor framework

    File Name: interceptor.py
    Description:

"""

import logging

from time import time

import grpc

from google.protobuf.json_format import MessageToJson

from .spec import mKeyRequestStringify
from .invocation_pb2 import Invocation

class EmptyServerInterceptor(object):
    """An empty server interceptor which does a little thing

    This interceptor should be a `example` interceptor or just used for a very simple usage.

    """
    def __init__(self, enableLogging = True):
        """Create a new EmptyServerInterceptor
        """
        self.enableLogging = enableLogging

    def __call__(self, request, context, methodDesc, handler):
        """Intercept
        """
        # Set logger
        if not hasattr(context, "logger"):
            if Invocation.current:
                context.logger = logging.getLogger("%s#[%s]" % (methodDesc.logger.name, Invocation.current.id or "N/A"))
            else:
                context.logger = methodDesc.logger
        # Write log
        if self.enableLogging:
            self.writeAccessLog(request, context, methodDesc, handler)
        # Call method
        try:
            if methodDesc.isServerStreamming:
                return self.callStreamMethod(request, context, methodDesc, handler)
            else:
                return self.callUnaryMethod(request, context, methodDesc, handler)
        except grpc.RpcError as error:
            # Forward grpc error
            context.logger.exception("gRPC error occurred")
            context.set_code(error.code())
            if error.code() == grpc.StatusCode.UNAVAILABLE:
                context.set_details("Service Unavailable")
            else:
                context.set_details(error.details())
            # Return
            return methodDesc.response()
        except Exception as error:
            # Error occurred
            context.logger.exception("Error occurred")
            if hasattr(error, "setGRPCContext"):
                getattr(error, "setGRPCContext")(context)
            else:
                # Set gRPC error context by interal error, we'll not expose any error details
                context.set_code(grpc.StatusCode.INTERNAL)
                context.set_details("Internal error")
            # Return
            return methodDesc.response()

    def callUnaryMethod(self, request, context, methodDesc, handler):
        """Call the unary method
        """
        # Call the handler
        rsp = None
        startTime = time()
        try:
            rsp = handler(request, context)
        except:
            raise
        finally:
            endTime = time()
            context.logger.info("Complete in %.4fs", endTime - startTime)
        # Check result
        if rsp is None:
            # This is not correct in python even when error occurred
            if self.enableLogging:
                context.logger.error("None type return value found, use empty response object instead")
            rsp = methodDesc.response()
        # Done
        return rsp

    def callStreamMethod(self, request, context, methodDesc, handler):
        """Call the stream method
        """
        # Call the handler
        startTime = time()
        try:
            for value in handler(request, context):
                yield value
        except:
            raise
        finally:
            endTime = time()
            context.logger.info("Complete in %.4fs", endTime - startTime)

    def writeAccessLog(self, request, context, methodDesc, handler):
        """Write access log
        """
        try:
            string = getattr(handler, mKeyRequestStringify)(request, context)
        except AttributeError:
            string = None
        except Exception as error:
            string = "Request stringify error: %s" % error
        # Write log
        if string:
            context.logger.info("Invoke: %s", string)
        else:
            context.logger.info("Invoke")
        # Write debug
        if context.logger.isEnabledFor(logging.DEBUG) and not methodDesc.descriptor.server_streaming:
            try:
                context.logger.debug("Request data:\n%s", MessageToJson(request))
            except Exception as error:
                context.logger.debug("Failed to dump request data, error: %s", error)

class EmptyClientInterceptor(object):
    """An empty client interceptor which does a little thing

    This interceptor should be a `example` interceptor or just used for a very simple usage.

    """
    def __init__(self, enableLogging = True):
        """Create a new EmptyClientInterceptor
        """
        self.enableLogging = enableLogging

    def __call__(self, handler, methodDesc, request, timeout = None, metadata = None, credentials = None):
        """Intercept
        """
         # Write log
        if self.enableLogging:
            self.writeCallLog(request, methodDesc, timeout, metadata, credentials)
        # Call the handler
        startTime = time()
        try:
            return handler(request, timeout, metadata, credentials)
        finally:
            endTime = time()
            if self.enableLogging:
                methodDesc.logger.info("Complete in %.4fs", endTime - startTime)

    def writeCallLog(self, request, methodDesc, timeout, metadata, credentials):
        """Write call log
        """
        # Write log
        methodDesc.logger.info("Call")
        # Write debug
        if methodDesc.logger.isEnabledFor(logging.DEBUG) and not methodDesc.descriptor.client_streaming:
            try:
                methodDesc.logger.debug("Request data:\n%s", MessageToJson(request))
                methodDesc.logger.debug("Request timeout: %s", timeout)
                methodDesc.logger.debug("Request metadata: %s", metadata)
                methodDesc.logger.debug("Request credentials: %s", credentials)
            except Exception as error:
                methodDesc.logger.debug("Failed to dump request data, error: %s", error)
