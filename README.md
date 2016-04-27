Fourchan - A 4Chan API Wrapper
==============================

Description
-----------

Fourchan is a simple 4Chan API wrapper written in Go that caches threads.

Installation
------------

You can `go get` it.

```
go get github.com/mikopits/fourchan
```

Or you can clone the latest Github repository.

```
git clone http://github.com/mikopits/fourchan
```

Examples
--------

Getting threads by page:

```go
vp := fourchan.NewBoard("vp", false)
page, _ := vp.GetThreadsByPage(1)
```

Getting threads by catalog:

```go
vp := fourchan.NewBoard("vp", false)
catalog, _ := vp.GetCatalog()
```
