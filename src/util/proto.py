# -*- coding: UTF-8 -*-

from google.protobuf.message import Message
from google.protobuf.any_pb2 import Any

from chameleon.api.media import Media

def getTypeUrl(msg:Message) -> str:
    type_url_prefix='type.googleapis.com/'
    return '{}{}'.format(type_url_prefix, msg.DESCRIPTOR.full_name)


def newImageMedia(id:str, orgUrl:str, largeUrl:str, mediumUrl:str, smallUrl:str, desc:str, isDefault:bool=False)->Media:
    img = Media.Image()
    img.id = id
    img.originalUrl = orgUrl
    img.largeUrl = largeUrl
    img.mediumUrl = mediumUrl
    img.smallUrl = smallUrl

    imgData = Any()
    imgData.Pack(img)
    media = Media()
    media.detail.CopyFrom(imgData)
    media.isDefault = isDefault
    media.text = desc or ""

    return media