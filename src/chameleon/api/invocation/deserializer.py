# encoding=utf8

""" An extension of protobuf json deserializer
    File Name: deserializer.py
    Description:

    NOTE:
        Currently we don't support deserialize timestamp / duration from non-standard date type in Any (Directly packed into anytype)
"""

from datetime import datetime, timedelta

from google.protobuf.json_format import _Parser

class Deserializer(_Parser):
    """Protobuf message deserializer
    """
    def __init__(self, ignoreUnknownFields = False):
        """Create a new Deserializer
        """
        super(Deserializer, self).__init__(ignore_unknown_fields = ignoreUnknownFields)

    def deserialize(self, value, message):
        """Deserialize
        """
        return self.ConvertMessage(value, message)

    def ConvertMessage(self, value, message):
        """Convert message
        """
        fullname = message.DESCRIPTOR.full_name
        if fullname == "google.protobuf.Timestamp":
            if isinstance(value, datetime):
                # Deserialize timestamp from datetime object
                message.seconds = int(value.strftime("%s"))
                message.nanos = int(float(value.strftime("%f")) * 10**3)
                return
        elif fullname == "google.protobuf.Duration":
            if isinstance(value, float):
                # Deserialize duration from float (in seconds)
                message.seconds = int(value)
                message.nanos = int((value - message.seconds) * 10**9)
                return
            elif isinstance(value, timedelta):
                # Deserialize duration from timedelta
                message.seconds = int(value.total_seconds())
                message.nanos = int(value.microseconds * 10**3)
        # Super
        super(Deserializer, self).ConvertMessage(value, message)
