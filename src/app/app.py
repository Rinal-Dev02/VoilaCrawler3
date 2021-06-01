#!/usr/bin/env python3
# -*- coding: UTF-8 -*-

import os
import sys
import time
import logging
import asyncio
import traceback
from queue import Empty, Queue,Full as FullException
from argparse import ArgumentParser
from typing import Callable
from google.protobuf.message import Message

import grpc
from google.protobuf.any_pb2 import Any
from google.protobuf.json_format import MessageToJson
from chameleon.smelter.v1.crawl import CrawlerNodeServicer, CrawlerRegisterStub, ConnectRequest, Error as RawError
from chameleon.smelter.v1.crawl.proxy import ProxyReliability, ReliabilityDefault, ReliabilityLow, ReliabilityMedium, ReliabilityHigh

from api.crawler import CrawlerNode
from crawler import Crawler, Error
from context import Context
from proxy import ProxyClient
from crawler import SupportedItemTypes
from crawler.context import Context, ReqIdKey, TargetTypeKey, TracingIdKey
from network.http.request import Request
from proxy.proxy import RequestOptions
from util.random import newRandomId

def getArgs():
    parser = ArgumentParser(prog="python -m <name>", description="spider service")

    subparsers = parser.add_subparsers(dest="cmd")
    serveParser = subparsers.add_parser("serve")
    serveParser.add_argument("--host", dest="host", default="0.0.0.0", help="gRPC serve host")
    serveParser.add_argument("--port", dest="port", type=int, default=6000, help="gRPC serve port", required=True)
    serveParser.add_argument("--crawlet-addr", dest="crawletAddr", help="crawlet gRPC server address", required=True)
    serveParser.add_argument("--proxy-addr", default=os.environ.get("VOILA_PROXY_URL"), dest="proxyAddr", help="proxy server address")
    serveParser.add_argument("--max-concurrency", dest="maxConcurrency", type=int, default=6, help="max grpc server concurrency")
    serveParser.add_argument("--session-addr", dest="sessionAddr", help="session server address")
    serveParser.add_argument("--debug", dest="debug", type=bool, default=False, nargs="?", const=True, help="enable debug")

    testParser = subparsers.add_parser("test")
    testParser.add_argument("--proxy-addr", default=os.environ.get("VOILA_PROXY_URL"), dest="proxyAddr", help="proxy server address")
    testParser.add_argument("--target", dest="target", help="target url to crawl")
    testParser.add_argument("--type", dest="types", nargs="+", default=None, choices=SupportedItemTypes, help="target type for crawl")
    testParser.add_argument("--level", dest="level", type=int, choices=[ReliabilityLow, ReliabilityMedium, ReliabilityHigh], default=None, help="proxy level")
    testParser.add_argument("--disable-proxy", dest="disableProxy", type=bool, default=None, nargs="?", const=True, help="disable proxy")
    testParser.add_argument("--enable-headless", dest="enableHeadless", type=bool, default=None, nargs="?", const=True,  help="enable headless")
    testParser.add_argument("--enable-session-init", dest="enableSessionInit", type=bool, default=None, nargs="?", const=True, help="enable session init")
    testParser.add_argument("--pretty", dest="pretty", type=bool, default=False, nargs="?", const=True, help="print result in pretty")
    testParser.add_argument("--debug", dest="debug", type=bool, default=False, nargs="?", const=True, help="enable debug")

    args = parser.parse_args()
    if args.proxyAddr:
        args.proxyAddr = os.environ["VOILA_PROXY_URL"]
    return args

class Application(object):
    """ Application """
    logger = logging.getLogger("Application")

    def __init__(self, args, newCrawler:Callable[[ProxyClient], Crawler]):
        level = logging.DEBUG if args.debug==True else logging.INFO
        logging.basicConfig(stream=sys.stdout, level=level, format = "%(asctime)s %(levelname)s %(name)s:%(message)s")

        self._args = args
        self._proxyClient = ProxyClient(args.proxyAddr)
        self._crawler = newCrawler(self._proxyClient)

    def _connect_sender(self, port:int):
        try:
            req = ConnectRequest.Ping()
            req.timestamp = int(time.time())
            req.id = self._crawler.ID()
            req.storeId = self._crawler.ID()
            req.version = self._crawler.Version()
            req.allowedDomains.MergeFrom(self._crawler.AllowedDomains())
            req.servePort = port
            data= Any()
            data.Pack(req)

            yield data

            start = time.time()
            defaultInterval = 4.5
            while True:
                end = time.time()
                interval = defaultInterval-(end-start)
                if interval > 0:
                    time.sleep(interval)
                start = time.time()
                req = ConnectRequest.Heartbeat()
                req.timestamp = int(time.time())
                data = Any()
                data.Pack(req)

                yield(data)
        except GeneratorExit:
            return
        except:
            print(traceback.format_exc())
            raise


    async def _serve(self):
        server = grpc.aio.server()

        crawler = CrawlerNode(self._proxyClient, self._crawler)
        CrawlerNodeServicer.addToServer(crawler, server)

        port = int(self._args.port or 6000)
        addr = "0.0.0.0:{}".format(port)
        self.logger.info("Start gRPC insecure server at %s", addr)
        server.add_insecure_port(addr)
        await server.start()

        try:
            # register to crawlet
            while True:
                try:
                    self.logger.info("connecting to crawlet %s", self._args.crawletAddr)
                    with grpc.insecure_channel(self._args.crawletAddr) as channel:
                        stub = CrawlerRegisterStub(channel)
                        iter = stub.Connect(self._connect_sender(port))
                        while True:
                            try:
                                next(iter)
                            except KeyboardInterrupt:
                                return
                            except:
                                self.logger.error("connect to crawlet failed, retry in 5 seconds")
                                time.sleep(5)
                                break
                except KeyboardInterrupt:
                    return
        except:
            self.logger.error(traceback.format_exc())
        finally:
            await server.stop(0)
            await server.wait_for_termination()

    def _local(self):
        """ run local functions """
        logging.StreamHandler(sys.stdout)

        ctx = Context(None, TracingIdKey, newRandomId())

        reqQueue = Queue(maxsize=1000)
        typs = self._args.types or list()
        reqFilter = set()
        if self._args.target:
            ctx = Context(ctx, TargetTypeKey, ",".join(typs))
            req = Request(ctx, "get", self._args.target)
            reqQueue.put(req)
            reqFilter.add(str(req.url))
        if reqQueue.empty():
            for r in self._crawler.NewTestRequest(ctx) or list():
                if not r:
                    continue
                nctx = ctx
                for (k,v) in r.context.values().items():
                    nctx = Context(nctx, k, v)
                r.context = nctx
                reqQueue.put(r)
                reqFilter.add(str(r.url))
        if reqQueue.empty():
            return

        while True:
            try:
                req = None
                try:
                    req = reqQueue.get(block=False)
                except Empty as e:
                    self.logger.info("no more requests")
                    return
                if not req:
                    return

                req.context = Context(req.context, ReqIdKey, newRandomId())
                opts = self._crawler.CrawlOptions(req.url)
                httpOpts = RequestOptions()
                httpOpts.enable_proxy = not self._args.disableProxy
                httpOpts.enable_headless = opts.enableHeadless
                httpOpts.enable_session_init = opts.enableSessionInit
                httpOpts.keep_session = opts.keepSession
                httpOpts.disable_cookie_jar = opts.disableCookieJar
                httpOpts.disable_redirect = opts.disableRedirect
                httpOpts.reliability = opts.reliability
                if self._args.enableHeadless is not None:
                    httpOpts.enable_headless = self._args.enableHeadless
                if self._args.enableSessionInit is not None:
                    httpOpts.enable_session_init = self._args.enableSessionInit
                if self._args.level is not None:
                    httpOpts.reliability = self._args.level
                for (k,v) in opts.headers.items():
                    req.headers.set(k, v)

                cookie = req.headers.get("cookie")
                for c in opts.cookies:
                    if c.path == "" or req.url.path.startswith(c.path):
                        val = "{}={}".format(c.name, c.value)
                        if not cookie:
                            cookie = val
                        else:
                            cookie = cookie + "; " + val
                if not cookie:
                    req.headers.set("cookie", cookie)
                try:
                    resp = self._proxyClient.do(req.context, req, httpOpts)
                    for e in self._crawler.Parse(req.context, resp) or list():
                        if not e:
                            continue
                        nctx, i = None, None
                        if not isinstance(e, tuple):
                            e = (e,)
                        for v in e:
                            if not v:
                                continue
                            if isinstance(v, Context):
                                if nctx == None:
                                    nctx = v
                            elif i == None:
                                i = v
                        if not i:
                            self.logger.error("got invalid yield")
                            continue
                        if not nctx:
                            if isinstance(i, Request):
                                nctx = i.context
                            else:
                                nctx = Context(ctx)
                        if isinstance(i, Request):
                            if str(i.url) in reqFilter:
                                continue
                            if not i.url.scheme:
                                i.url.scheme = "https"
                            if not i.url.host:
                                i.url.host = req.url.host

                            try:
                                reqQueue.put_nowait(i)
                                reqFilter.add(str(i.url))
                            except FullException as e:
                                self.logger.error("queue is full, ignored the request")
                        elif isinstance(i, Error) or isinstance(i, RawError):
                            self.logger.error("got message error {}", str(i))
                        elif isinstance(i, Message):
                            data = MessageToJson(i, including_default_value_fields=True, indent=4)
                            self.logger.info("data: %s", data)
                        else:
                            raise ValueError("unsupported data type {}".format(i))
                except KeyboardInterrupt:
                    raise
            except KeyboardInterrupt:
                return
            except:
                self.logger.info(traceback.format_exc())
                return

    def run(self):
        if self._args.cmd == "serve":
            asyncio.run(self._serve())
        elif self._args.cmd == "test":
            self._local()
