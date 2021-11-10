# Simple HTTP dummy images server

URL format:

    http://<host>:<port>/<text>/<width>x<height>.<format>

where:

* `host` - server host or IP address
* `port` - server port (default: 8080, can be set by option `-port`)
* `text` - text for image. Can be skipped, using image resolution by default. Have 2 predefined values:
    * `timestamp` - current UNIX timestamp in milliseconds
    * `datetime` - current date and time
* `width` - image width in pixels
* `height` - image height in pixels
* `format` - one of formats: png, jpg, jpeg, gif

Examples:

* http://localhost:8080/100x100.png
* http://localhost:8080/timestamp/640x480.gif
* http://localhost:8080/datetime/1080x1080.jpg
* http://localhost:8080/some%20text/1920x1080.jpg
