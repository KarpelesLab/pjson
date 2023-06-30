[![GoDoc](https://godoc.org/github.com/KarpelesLab/pjson?status.svg)](https://godoc.org/github.com/KarpelesLab/pjson)

# encoding/json fork

This is a fork of `encoding/json` with the following features added:

* Context support: MarshalContext() accepts a context that can be accessed by implementing a variant of MarshalJSON()
* Groups: it is possible to have methods returning groups of values

## Context

Objects can now implement a `MarshalContextJSON(context.Context)` method that will be called over the usual `MarshalJSON()` with
the context of the encoding. `Marshal(v)` will use `context.Background()` by default, and the new `MarshalContext(ctx, v)` receives
a context that will be passed to all objects encoded in the process.

## Grouping

Objects can implement a `GroupMarshalerJSON(ctx context.Context, st *pjson.GroupState) ([]byte, error)` method that will be
called upon marshaling and can use GroupState to specify it needs to gather specific data in batches before rending is possible.

This can be useful when data returned from a json encoding is taken from a database, and many objects access elements from a given
list. These can be grouped in a single SELECT (or similar) from the database, greatly reducing the time needed for render.
