annolityx
=========
Annolityx allows you to capture and annotate events.  It is designed to integrate with dashboarding engines.


## Getting started

To get the source, build and install annolityx:

    go get github.com/metrilyx/annolityx

On success, you can find the binary at **$GOPATH/bin/annolityx**.  The service can now be started by issueing the following command:

    annolityx -c $GOPATH/github.com/metrilyx/annolityx/conf/annolityx.toml -webroot $GOPATH/github.com/metrilyx/annolityx/webroot -l info    

You can also provide that webroot path in the configuration file.  Once the service has been started you can view the UI at **http://localhost:9898/**

Continue on to the configuration section to customize various components of the application.


## Configuration

- Configuration file: **$GOPATH/github.com/metrilyx/annolityx/conf/annolityx.toml**.  
- Web UI directory: **$GOPATH/github.com/metrilyx/annolityx/webroot**.


## Usage
The following enpoints are available:

    /api/annotations (GET, POST)
    /api/types       (GET)

### Annotation types
Each annotation as an associated type used for grouping and categorizing.  The type must exist before an annotation can be use it.

To retrieve a list of available types:

**Request:**

    curl http://localhost:9898/api/types

**Response:**
A list of annotation types as JSON objects.

    [
        {
            "id": "alarm",
            "name": "Alarm",
            "metadata": {
                "color": "#ff0000"
            }
        }
    ]

### Annotations

#### Schema
An annotation schema is layed out as follows:

* posted_timestamp: time in epoch when event was submitted to the system
* timestamp       : time in epoch when event occurred
* data            : single level json object with arbitrary user data
* tags            : single level json object with arbitrary user tags used for filtering and searching
* message         : user message string

#### Querying
To retrieve annotations a GET request must be made:

**Request:**

    curl -XGET "http://localhost:9898/api/annotations?tags=host:foo,dc:dc1&types=alarm,release&start=2014.11.01-00:00:00"

or the same request with a GET body:

    curl -XGET http://localhost:9898/api/annotations -d '{
        "tags": {
            "host": "foo",
            "dc": "dc1"
        },
        "types": ["alarm", "release"],
        "start": "2014.11.01-00:00:00"
    }'


**Response:**
A list of JSON objects.

    [
        {
            "tags": {"host":"foo"},
            "type": "alarm",
            "timestamp": 1418531584.234390,
            "posted_timestamp": 1418531584.234390,
            "data": {
                "name": "test"
            },
            "message": "Some message"
        }
    ]

| Field | Description | Required | Type |
|-------|-------------| ----------|------|
|**tags**|Tags to use to filter events.  These work as an **and**| **Yes**| dict|
|**types**|1 or more event types.  These work as an **or**|**Yes**| array |
|**start**|Start time in epoch seconds or string with the appropriate format. |**Yes**| float or string (YYYY.MM.DD-hh:mm:ss) |
|**end**|End time in epoch seconds or string with the appropriate format. |No| float or string (YYYY.MM.DD-hh:mm:ss) |


#### Annotating
To submit an annotation make a POST request to the same endpoint as follows:

**Request:**

    curl -XPOST http://localhost:9898/api/annotations -d '{
        "tags": {
            "host": "foo",
            "dc": "dc1"
        },
        "type": "alarm",
        "message": "Memory usage at 90%",
        "timestamp": "2014.11.01-00:00:00",
        "data": {
            "host": "foo.bar.org",
            "time": "2014/12/23-01:00:01",
            "contact": "admin@foo.bar.org"
        }
    }

**Response:**

    {
        "id": "... sha1 sum...",
        "posted_timestamp": 1418531584.234390,
        "tags": {
            "host": "foo",
            "dc": "dc1"
        },
        "type": "alarm",
        "message": "Memory usage at 90%",
        "timestamp": 1418531584.234322,
        "data": {
            "host": "foo.bar.org",
            "time": "2014/12/23-01:00:01",
            "contact": "admin@foo.bar.org"
        }
    }


| Field | Description | Example | Required | Type |
|-------|-------------|---------|----------|------|
| **timestamp** | Epoch time in **seconds** (UTC).  If not provided the current time is used. | 1408129158 (Aug 15 11:59:22 2014) | No | float or string (YYYY.MM.DD-hh:mm:ss) |
| **type** | A pre-defined event type.  A list of event types can be found at the /api/types endpoint. | Maintenance | **Yes** | string |
| **message** | This is the string used when hovering over the event on the graph. | "Scheduled Network Maintenance"| **Yes** | string |
| **tags** | Any arbitrary tags that can be used later for searching/filtering. | {"host":"foo.bar.com","severity":"Warning"}| **Yes** | dict |
| **data** | This can be any arbitrary JSON data.  It must be a single level JSON structure. This is the data used as details which are shown when clicking on the event| {"Priority": "P1", "On Call": "Jon Doe", "Contact Email": "Jon.Doe@bar.com" }| No | dict |
