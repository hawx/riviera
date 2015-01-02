# Riviera

A [river.js][] generator written in Go.

``` bash
$ go get github.com/hawx/riviera
```

This serves a single file `/river.js` showing a river for the feeds listed in
`subscriptions.xml`:

``` bash
$ riviera --opml subscriptions.xml
XXXX/XX/XX XX:XX:XX web.go serving 0.0.0.0:8080
```

It will pull every feed, and then keep them updated.


## Admin routes

There are routes for administering a running river, these must be enabled with
the `--with-admin` flag. They have no authentication so should be hidden from
the public. There is an admin interface that uses these at
[riviera-admin][]. For now they are,

```
/-/list
/-/subscribe?url=...
/-/unsubscribe?url=...
```

These may change.


## Reading the feed

You will now need a front-end to consume the river, I currently use [rivelin][]
but did use [necolas/newsriver-ui][newsriver-ui] before. In either case you will
need to follow the instructions given and put the correct url to the `river.js`
file generated. For instance, if you ran riviera at `http://example.com` the
file is generated at `http://example.com/river.js`.

[river.js]:      http://riverjs.org
[riviera-admin]: https://github.com/hawx/riviera-admin
[rivelin]:       https://github.com/hawx/rivelin
[newsriver-ui]:  https://github.com/necolas/newsriver-ui
