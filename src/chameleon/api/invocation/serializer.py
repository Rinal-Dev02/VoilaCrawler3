# encoding=utf8

""" An extension of protobuf json serializer

    File Name: serializer.py
    Description:

    NOTE:

        Currently we don't support serialize timestamp / duration to non-standard date type in Any (Directly packed into anytype)

"""

from datetime import datetime, timedelta

from google.protobuf.json_format import _Printer

OptionsTimestamp2Datetime       = "ts2dt"
OptionsDuration2Timedelta       = "dr2td"
OptionsDuration2FloatSeconds    = "dr2floats"

_SecondsPerDay = 3600 * 24

class Serializer(_Printer):
    """Protobuf message serializer
    """
    def __init__(self, includingDefaultValueFields = False, preservingProtoFieldName = False, options = None):
        """Create a new Serializer
        """
        self.options = options
        self.typefuncs = {}
        if options:
            for option in options:
                if option == OptionsTimestamp2Datetime:
                    self.typefuncs["google.protobuf.Timestamp"] = self.timestamp2datetime
                elif option == OptionsDuration2Timedelta:
                    self.typefuncs["google.protobuf.Duration"] = self.duration2timedelta
                elif option == OptionsDuration2FloatSeconds:
                    self.typefuncs["google.protobuf.Duration"] = self.duration2floatseconds
        # Super
        super(Serializer, self).__init__(includingDefaultValueFields, preservingProtoFieldName)

    def serialize(self, message):
        """Serializer the message
        """
        return self._MessageToJsonObject(message)

    def _MessageToJsonObject(self, message):
        """Message to json object
        """
        typefunc = self.typefuncs.get(message.DESCRIPTOR.full_name)
        if typefunc:
            return typefunc(message)
        return super(Serializer, self)._MessageToJsonObject(message)

    # -*- ------------------------------ Serialization function ------------------------------ -*-

    def timestamp2datetime(self, message):
        """Convert google.protobuf.Timestamp to datetime object
        """
        return datetime.utcfromtimestamp(message.seconds + message.nanos / float(10**9))

    def duration2timedelta(self, message):
        """Convert google.protobuf.Duration to timedelta object
        """
        return timedelta(int(message.seconds / _SecondsPerDay), int(message.seconds % _SecondsPerDay), int(message.nanos // 10**6), int((message.nanos % 10**6) // 10**3))

    def duration2floatseconds(self, message):
        """Convert google.protobuf.Duration to float in second
        """
        return message.seconds + message.nanos / float(10**9)
