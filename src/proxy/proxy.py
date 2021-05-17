# -*- coding: UTF-8 -*-

import logging
import gzip
from typing import Type

from requests.api import head 
import brotli
import requests
import traceback
from google.protobuf import json_format

from context import Context
from chameleon.api.http.data_pb2 import ListValue
from network.http.header  import Header
from network.http.request import Request,Response
from crawler.context import TracingIdKey,JobIdKey,ReqIdKey,StoreIdKey,TargetTypeKey
from chameleon.smelter.v1.crawl.proxy import Request as ProxyRequest,Response as ProxyResponse,ProxyReliability,ReliabilityDefault,ReliabilityLow,ReliabilityMedium,ReliabilityHigh,ReliabilityIntelligent
from util import newRandomId

StringList = list[str]

class RequestOptions(object):
    """ Request Options  """

    def __init__(self, enable_proxy:bool=True, enable_headless:bool=False, js_wait_duration:int=0, enable_session_init:bool=False,
        keep_session:bool=False, disable_cookie_jar:bool=False, disable_redirect:bool=False, relibility:ProxyReliability=ReliabilityDefault, filter_keys:list[str]=None):

        # enable proxy
        self._enable_proxy = enable_proxy
        # enable headless
        self._enable_headless = enable_headless
        # default 6 seconds
        self._js_wait_duration = js_wait_duration or 0
        # enable session init
        self._enable_session_init = enable_session_init
        self._keep_session = keep_session
        self._disable_cookie_jar = disable_cookie_jar
        self._disable_redirect = disable_redirect
        self._reliability = ReliabilityDefault
        self._request_filter_keys = filter_keys or list()

    @property
    def enable_proxy(self)->bool:
        return self._enable_proxy

    @enable_proxy.setter
    def enable_proxy(self, val:bool):
        self._enable_proxy = val

    @property
    def enable_headless(self)->bool:
        return self._enable_headless

    @enable_headless.setter
    def enable_headless(self, val:bool):
        self._enable_headless = val

    @property
    def js_wait_duration(self)->bool:
        return self._js_wait_duration

    @js_wait_duration.setter
    def js_wait_duration(self, val):
        if val < 0:
            raise ValueError("duration must >=0")
        self._js_wait_duration = val

    @property
    def enable_session_init(self)->bool:
        return self._enable_session_init

    @enable_session_init.setter
    def enable_session_init(self, val:bool):
        self._enable_session_init = val

    @property
    def keep_session(self)->bool:
        return self._keep_session

    @keep_session.setter
    def keep_session(self, val:bool):
        self._keep_session = val

    @property
    def disable_cookie_jar(self)->bool:
        return self._disable_cookie_jar

    @disable_cookie_jar.setter
    def disable_cookie_jar(self, val:bool):
        self._disable_cookie_jar = val

    @property
    def disable_redirect(self)->bool:
        return self._disable_redirect

    @disable_redirect.setter
    def disable_redirect(self, val:bool):
        self._disable_redirect = val

    @property
    def reliability(self)->bool:
        return self._reliability

    @reliability.setter
    def reliability(self, val:ProxyReliability):
        self._reliability = val

    @property
    def request_filter_keys(self)->StringList:
        return self._request_filter_keys

    @request_filter_keys.setter
    def request_filter_keys(self, val:StringList):
        self._request_filter_keys = val

class ProxyClient(object):
    """ Client """

    logger = logging.getLogger("ProxyClient")

    def __init__(self, proxy_addr:str):
        if not proxy_addr:
            raise ValueError("invalid proxy address")
        self._proxy_addr = proxy_addr

    def do(self, ctx:Context, r:Request, options:RequestOptions)->Response:
        """ do http request """

        self.logger.info("access %s", str(r.url))

        _resp = None
        try:
            req = ProxyClient.buildRequest(ctx, r, options)
            reqData = json_format.MessageToDict(req)
            _resp = requests.post(self._proxy_addr, json=reqData)
            if _resp.status_code != 200:
                # proxy response
                raise ValueError("request failed with status {}".format(_resp.status_code))
            resp = ProxyResponse()
            # not ignored the unknown fields
            json_format.Parse(_resp.content, resp)
            return ProxyClient.buildResponse(ctx, resp)
        except Exception as e:
            ProxyClient.logger.error(traceback.format_exc())
            raise
        finally:
            if _resp:
                _resp.close()

    @staticmethod
    def buildRequest(ctx:Context, r:Request, opts:RequestOptions)->ProxyRequest:
        if not r:
            raise ValueError("invalid request")
        req = ProxyRequest()
        req.tracingId = ctx.get_str(TracingIdKey)
        req.jobId = ctx.get_str(JobIdKey)
        req.reqId = ctx.get_str(ReqIdKey)
        if not req.reqId:
            req.reqId = newRandomId()
        req.method = r.method
        req.url    = str(r.url)
        req.body   = r.body or bytes()
        for (k,v) in (r.headers).values:
            req.headers[k].values.MergeFrom(v)
        req.options.enableProxy = opts.enable_proxy
        req.options.enableHeadless = opts.enable_headless
        req.options.enableSessionInit = opts.enable_session_init
        req.options.keepSession = opts.keep_session
        req.options.disableCookieJar = opts.disable_cookie_jar
        req.options.disableRedirect = opts.disable_redirect
        req.options.reliability = opts.reliability
        req.options.maxTtlPerRequest = 5*60
        req.options.jsWaitDuration = opts.js_wait_duration
        req.options.requestFilterKeys.MergeFrom(opts.request_filter_keys)

        return req

    @staticmethod
    def buildResponse(ctx:Context, r:ProxyResponse)->Response:
        def builder(ctx:Context, r:ProxyResponse, isSub:bool=False)->Response:
            if not r:
                return None
            header = Header()
            for (key,listval) in r.headers.items():
                for val in listval.values:
                    header.add(key, val)

            body = r.body
            if not isSub and len(r.body) > 0:
                if "gzip" in header.get("content-encoding"):
                    body = gzip.decompress(r.body)
                    header.delete("content-encoding")
                elif "br" in header.get("content-encoding"):
                    body = brotli.decompress(r.body)
                    header.delete("content-encoding")
            subresp = None
            if r.request.HasField("response"):
                subresp = builder(r.request.response, True)
            reqHeader = Header()
            for (key,listval) in r.headers.items():
                for val in listval.values:
                    reqHeader.add(key, val)
            req = Request(ctx, r.request.method, r.request.url, header=reqHeader, resp=subresp)
            return Response(r.statusCode, header, body, req)
        return builder(ctx, r)