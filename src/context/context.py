# -*- coding: UTF-8 -*-

class Context:
    """ Context is an node list """

    def __init__(self, parent, key=None, value=None):
        """ init context """
        self._parent = parent
        self._key = key
        self._value = value

    def get(self, key):
        """ get value with raw type """
        if self._key == key:
            return self._value
        if not self._parent or not isinstance(self._parent, Context): return None
        return self._parent.get(key)

    def get_str(self, key):
        """ get str value """
        val = self.get(key)
        if not val:
            return ""
        return str(val)
    
    def get_int(self, key):
        """ get int value """
        val = self.get(key)
        if not val:
            return 0
        try:
            n = int(val)
            return n
        except:
            return 0

    def values(self):
        """ value """
        vals = None
        if not self._parent and not isinstance(self._parent, Context):
            vals = dict()
        else:
           vals = self._parent.values()
        if self._key != None:
            vals[self._key] = self._value
        return vals