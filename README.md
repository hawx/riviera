# Riviera

A river-of-news style feed aggregator, using [river.js][].

``` bash
$ go get hawx.me/code/riviera
$ riviera --boltdb ./mydb feeds.xml
...
```

Riviera expects a list of feeds to subscribe to be passed. This is given in
[OPML subscription list][opml] format. An example,

``` xml
<?xml version="1.0" encoding="UTF-8"?>
<opml version="1.1">
  <head>
    <title>Subscriptions</title>
  </head>
  <body>
    <outline type="rss" xmlUrl="http://feeds.bbci.co.uk/news/uk/rss.xml"></outline>
    <outline type="rss" xmlUrl="http://feeds2.feedburner.com/TheAwl"></outline>
    <outline type="rss" xmlUrl="http://feeds.kottke.org/main"></outline>
  </body>
</opml>
```

By default an in-memory database is used, it is more useful to use the
`--boltdb` option to create/open a database on disk.

The riverjs document is served at `/river`, a set of metadata listing the feeds
subscribed to and recent fetcher activity is served at `/river/meta`.

See `riviera --help` for a full list of options.


## Reading

The output from riviera should be compatible with any application that can read
riverjs format feeds. I currently use [rivelin][] to read my feeds, but I have
in the past used [necolas/newsriver-ui][newsriver-ui].

In either case you will need to follow the instructions given and put the
correct url to the generated file, remembering that it will be
`http://example.com/river/` not `http://example.com/`.


## Subscribing / Unsubscribing

Riviera watches the file containing your subscription list for changes and will
attempt to update the feeds it is subscribed to based on changes to it.

That said it isn't the best experience to have to modify a file on a server to
subscribe to a feed. Using [riviera-admin][] provides a simple admin interface,
including a bookmarklet to subscribe to a site's feed.


[river.js]:      http://riverjs.org
[riviera-admin]: https://github.com/hawx/riviera-admin
[rivelin]:       https://github.com/hawx/rivelin
[newsriver-ui]:  https://github.com/necolas/newsriver-ui
[opml]:          http://dev.opml.org/spec2.html#subscriptionLists
