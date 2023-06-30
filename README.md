# encoding/json fork

This is a fork of `encoding/json` with the following features added:

* Context support: MarshalContext() accepts a context that can be accessed by implementing a variant of MarshalJSON()
* Groups: it is possible to have methods returning groups of values
