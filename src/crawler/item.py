# -*- coding: UTF-8 -*-

import time
from google.protobuf.any_pb2 import Any
from google.protobuf.message import Message

from context import Context
from chameleon.smelter.v1.crawl import Item
from chameleon.smelter.v1.crawl.item import Product, Tiktok, Youtube
from .context import TracingIdKey,JobIdKey,ReqIdKey,StoreIdKey,IndexKey
from util.proto import getTypeUrl

SupportedItemTypes = [getTypeUrl(t) for t in [Product(),Tiktok.Author(),Tiktok.Item(),Youtube.Channel(),Youtube.Video()]]

class Item(object):
    """ Error """

    def __init__(self, msg:Message):
        self._message = msg
        self._timestamp = int(time.time()*1000)

    @property
    def message(self):
        """ message """
        return self._message

    def encode(self, ctx:Context)->Item:
        ctx = ctx or Context()

        msg = Any()
        msg.Pack(self._message)

        item = Item()
        item.tracingId = ctx.get_str(TracingIdKey)
        item.jobId = ctx.get_str(JobIdKey)
        item.storeId = ctx.get_str(StoreIdKey)
        item.reqId = ctx.get_str(ReqIdKey)
        item.index = ctx.get_int(IndexKey)
        item.data.CopyFrom(msg)
        item.timestamp = self._timestamp

        return item