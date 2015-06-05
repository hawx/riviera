# Riviera

A [river.js][] generator written in Go.

``` bash
$ go get hawx.me/code/riviera
$ riviera --boltdb ./mydb
XXXX/XX/XX XX:XX:XX listening on port :8080
...
```

The app serves two files by default:

- `/river.js`
- `/subscriptions.opml`

If the `--with-admin` flag is passed, the following routes are also exposed:

- `/-/list`
- `/-/subscribe`
- `/-/unsubscribe`

By default an in memory database will be used, so to persist data specify a path
with the `--boltdb` option.


## Importing existing subscriptions

You can import an [opml][] file containing a list of subscriptions by
using the `--opml` flag. For example, if my list is stored in
`subscriptions.xml`:

``` bash
$ riviera --opml subscriptions.xml
XXXX/XX/XX XX:XX:XX imported subscriptions.xml
$ riviera
XXXX/XX/XX XX:XX:XX listening on port :8080
...
```

## Admin

The admin routes have no authentication, so should be hidden from the
public. This is the reason they must be explicitly enabled. There is an admin
interface that uses these at [riviera-admin][]. Though the routes may change
slightly in the future, it is unlikely from this point on that backwards
incompatible changes will be made. So here is a short description:

<dl>
  <dt><code>/-/list</code></dt>
  <dd>List returns a json representation of the feeds subscribed to.</dd>
  <dt><code>/-/subscribe?url=...</code></dt>
  <dd>Subscribes to the url passed as the parameter. Riviera will then
  immediately fetch the feed and add the latest items to the river.</dd>
  <dt><code>/unsubscribe?url=...</code></dt>
  <dd>Unsubscribes from the url passed. This does not remove past items so you
  will potentially have to wait until the cut-off period has been exceeded
  before all items disappear.</dd>
</dl>

## Reading

You will now need a front-end to consume the river, I currently use [rivelin][]
but did use [necolas/newsriver-ui][newsriver-ui] before. In either case you will
need to follow the instructions given and put the correct url to the `river.js`
file generated. For instance, if you ran riviera at `http://example.com` the
file is generated at `http://example.com/river.js`.


[river.js]:      http://riverjs.org
[riviera-admin]: https://github.com/hawx/riviera-admin
[rivelin]:       https://github.com/hawx/rivelin
[newsriver-ui]:  https://github.com/necolas/newsriver-ui
[opml]:          http://en.wikipedia.org/wiki/OPML
