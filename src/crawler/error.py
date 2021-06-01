# -*- coding: UTF-8 -*-

import time
from chameleon.smelter.v1.crawl import Error as RawError

from context import Context
from .context import TracingIdKey,JobIdKey,ReqIdKey,StoreIdKey

from errors import Code,OK,Cancelled,Unknown,InvalidArgument,DeadlineExceeded,NotFound,AlreadyExists,PermissionDenied,Unauthenticated,ResourceExhausted,FailedPrecondition,Aborted,OutOfRange,Unimplemented,Internal,Unavailable,DataLoss

class Error(Exception):
    """ Error """

    def __init__(self, msg, code:Code=OK):
        super(Error, self).__init__()

        self._timestamp = int(time.time()*1000)
        if isinstance(msg, str):
            self._message = msg
            self._code = code
        elif isinstance(msg, Exception):
            self._message = str(msg)
            self._code = code or Internal
        elif isinstance(msg, Error):
            self._message = msg.errMsg
            self._code = msg.code
        else:
            raise TypeError("unsupported error message type")

    @property
    def code(self):
        """ code """
        return self._code

    @property
    def message(self):
        """ message """
        return self._message

    def encode(self, ctx)->RawError:
        """ Error """
        ctx = ctx or Context()

        e = Error()
        e.tracingId = ctx.get_str(TracingIdKey)
        e.jobId = ctx.get_str(JobIdKey)
        e.storeId = ctx.get_str(StoreIdKey)
        e.reqId = ctx.get_str(ReqIdKey)
        e.code = self._code
        e.message = self._message
        e.timestamp = self._timestamp

        return e

ErrAbort = Error("abort the progress", code=Aborted)
ErrUnsupportedPath = Error("unsupported parse path", code=Unimplemented)
