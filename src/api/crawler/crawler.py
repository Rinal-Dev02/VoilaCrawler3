# -*- coding: UTF-8 -*-

import time
import logging
import traceback

from google.protobuf import any_pb2
from google.protobuf.any_pb2 import Any
from google.protobuf.empty_pb2 import Empty
from google.protobuf.message import Message

from context import Context
from crawler import Crawler,Error,Item,TracingIdKey,JobIdKey,ReqIdKey,StoreIdKey,TargetTypeKey
from errors.code_pb2 import Internal
from proxy import ProxyClient, RequestOptions
from network.http import Request,Header

from chameleon.api import http
from chameleon.smelter.v1.crawl import Error as RawError,Request as RawRequest,CrawlerNodeServicer,VersionResponse,CrawlerOptionsResponse,AllowedDomainsResponse,CanonicalUrlResponse

class CrawlerNode(CrawlerNodeServicer):
    logger = logging.getLogger("crawler.servicer.CrawlerNodeServicer")

    def __init__(self, httpClient:ProxyClient, crawler:Crawler):
        if not httpClient:
            raise ValueError("invalid proxy http client")
        if not crawler:
            raise ValueError("invalid crawler")
        self._httpClient = httpClient
        self._crawler = crawler

    async def Version(self, req, context):
        v = self._crawler.Version()
        resp = VersionResponse()
        resp.version = v
        return resp

    async def CrawlerOptions(self, req, context):
        opts = self._crawler.CrawlerOptions(req.url)
        resp = CrawlerOptionsResponse()
        resp.data = opts

        # TODO: type convert
        return resp

    async def AllowedDomains(self, req, context):
        domains = self._crawler.AllowedDomains()
        resp = AllowedDomainsResponse()
        resp.data = domains
        return resp

    async def CanonicalUrl(self, req, context):
        u = self._crawler.CanonicalUrl(req.url)
        resp = CanonicalUrlResponse()
        resp.data.url = u
        return resp

    @staticmethod
    def buildReq(ctx:Context, r:RawRequest)->Request:
        header = Header()
        for (k,v) in r.customHeaders:
            if k.lower() == "cookie":
                continue
            header.set(k,v)
        cookie = ""
        cookieFilter = set()
        for c in r.customCookies:
            if c.name in cookieFilter:
                continue
            cookieFilter.add(c.name)

            cv = c.name+"="+c.value
            if cookie == "":
                cookie = cv
            else:
                cookie = cookie + "; " + cv
        header.set("cookie", cookie)

        return Request(ctx, r.method, r.url, r.body, header)

    async def Parse(self, rawreq:RawRequest, context):
        """ Parse """

        ctx = Context(context)
        for (k,v) in (rawreq.sharingData or dict()).items():
            ctx = Context(ctx, k, v)
        ctx = Context(ctx, TracingIdKey, rawreq.tracingId)
        ctx = Context(ctx, JobIdKey, rawreq.jobId)
        ctx = Context(ctx, ReqIdKey, rawreq.reqId)
        ctx = Context(ctx, StoreIdKey, rawreq.storeId)
        ctx = Context(ctx, TargetTypeKey, ",".join(rawreq.options.targetTypes or list()))

        newReq = CrawlerNode.buildReq(ctx, rawreq)

        copts = self._crawler.CrawlOptions(newReq.url)
        opts = RequestOptions()
        opts.enable_proxy = not rawreq.options.diableProxy
        opts.enable_headless = copts.enableHeadless
        opts.enable_session_init = copts.enableSessionInit
        opts.keep_session = copts.keepSession
        opts.disable_cookie_jar = copts.disableCookieJar
        opts.disable_redirect = copts.disableRedirect
        # TODO: use dynamic config
        opts.reliability = copts.reliability
        resp = self._httpClient.do(newReq, opts)

        if not resp.body:
            raise ValueError("no response got")

        rawurl = newReq.url
        try:
            for e in self._crawler.Parse(ctx, resp):
                nctx, i = None, None
                if not isinstance(e, tuple):
                    e = tuple(e)
                for v in e:
                    if not v:
                        continue
                    if isinstance(v, Context):
                        if nctx == None:
                            nctx = v
                    elif i == None:
                        i = v
                if not i:
                    CrawlerNode.logger.error("got invalid yield")
                    continue
                if not nctx:
                    if isinstance(i, Request):
                        nctx = i.context
                    else:
                        nctx = Context(ctx)

                for key in [TracingIdKey, JobIdKey, StoreIdKey, ReqIdKey, TargetTypeKey]:
                    if not nctx.get(key):
                        nctx = Context(nctx, key, ctx.get(key))

                if isinstance(i, Request):
                    u = i.url
                    if u.host == "":
                        u.scheme = rawurl.scheme
                        u.host = rawurl.host
                    elif u.scheme != "http" and u.scheme != "https":
                        u.scheme = rawurl.scheme
                    if not i.header.get("referer"):
                        i.header.set("referer", str(rawurl))

                    subreq = RawRequest()
                    subreq.tracingId = rawreq.tracingId
                    subreq.jobId = rawreq.jobId
                    subreq.reqId = rawreq.reqId
                    subreq.storeId = rawreq.storeId
                    subreq.url = str(u)
                    subreq.method = i.method
                    subreq.body = i.body
                    subreq.parent.CopyFrom(rawreq)
                    subreq.customHeaders.CopyFrom(rawreq.customHeaders)
                    subreq.customCookies.CopyFrom(rawreq.customCookies)
                    subreq.options.CopyFrom(rawreq.options)
                    subreq.sharigData.CopyFrom(rawreq.sharingData)
                    for (k,v) in i.context().values().items():
                        if k in set([TracingIdKey, JobIdKey, ReqIdKey, StoreIdKey, TargetTypeKey]):
                            continue
                        subreq.sharingData[k] = v
                    ret = Any()
                    ret.Pack(subreq)
                    yield ret
                elif isinstance(i, RawError) or isinstance(i, Error):
                    e = i
                    if isinstance(i, RawError):
                        e = Error(i)
                    ret = Any()
                    ret.Pack(e.encode(nctx))
                    yield(ret)
                elif isinstance(i, Message):
                    try:
                        item = Item(i)

                        ret = Any()
                        ret.Pack(item.encode(nctx))
                        yield ret
                    except Exception as e:
                        e = Error(e, Internal)
                        ret = Any()
                        ret.Pack(e.encode(nctx))
                        yield(ret)
                else:
                    CrawlerNode.logger.error("unsupported yield data type {}".format(type(i)))
                    e = Error("unsupported yield data type {}".format(type(i)))
                    ret = Any()
                    ret.Pack(e.encode(nctx))
                    yield(ret)
        except Error as err:
            e = err.encode(ctx)
            ret = Any()
            ret.Pack(e)
            yield ret
        except Exception:
            msg = traceback.format_exc()
            e = Error(msg, code=Internal).encode(ctx)
            ret = Any()
            ret.Pack(e)
            yield ret
            
