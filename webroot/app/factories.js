var appFactories = angular.module("app.factories", [])
.factory("WebSocketManager", [ 'AnnolityxConfig', function(AnnolityxConfig) {

    var WebSocketManager = function(cb, messageFilter) {

        var t = this;

        var wsock;
        //var uri = AnnolityxConfig.websocket.url;
        
        var maxRetries = 3;
        var retryCount = 0;
        var retryInterval = 10000;
        
        var msgCallback = cb;
        var messageFilter = messageFilter;

        function sendMessage(data) {
            wsock.send(JSON.stringify(data));
        }

        function onWsOpen(evt) {
            console.log("Connection open. Extentions: [ '" + wsock.extensions + "' ]");
            retryCount = 0;
            sendMessage(messageFilter);
        }

        function onWsClose(evt) {
            console.log('Connection closed', evt);
            wsock = null;
            
            if(retryCount < maxRetries) {

                console.log('Reconnecting in 5sec...');

                setTimeout(function() {
                    connect();
                }, retryInterval);

                retryCount++;
            } else {
                console.log('Max retries exceeded!');
            }
        }

        function msgErrback(e) {
            console.error('Subscriber error:', evt.data);
            console.warn(e);
        }

        function onWsMessage(evt) {
            var data;
            try {
                data = JSON.parse(evt.data);
            } catch(e) {
                msgErrback(e)
                return;
            }
            if(data.error !== undefined) {
                msgErrback(evt);
            } else {
                msgCallback(data);
            }
        }

        function connect() {
            wsock = new WebSocket(AnnolityxConfig.websocket.url);
            wsock.addEventListener('open', onWsOpen);
            wsock.addEventListener('message', onWsMessage);
            wsock.addEventListener('close', onWsClose);
        }

        function _initialize() {
            
            for(var k in messageFilter) {
                
                if(k !== 'tags' && k !== 'types') 
                    delete messageFilter[k];
            }

            t.sendMessage = sendMessage
            t.connect = connect;
        }

        _initialize();
    };

    return (WebSocketManager);
}])
.factory("TimeWindowManager", [ '$location', '$routeParams', function($location, $routeParams) {

    var TimeWindowManager = function(scope) {

        var t = this;
        var DATE_TIME_FORMAT = "YYYY.MM.DD-HH:mm:ss"

        var timeDimension = {
            types: {
                relative: { start: '8h' },
                absolute: { start: new Date((new Date()).getTime()-(8*3600000)), end: new Date() }
            },
            activeType: 'relative'
        };

        function parseTime(timeObj) {
            if(timeObj.start !== '') {
                var matches = timeObj.start.match(/(\d+[s|m|h|d|w])-ago/)
                if(matches) {
                    timeDimension.types.relative.start = matches[1];
                    timeDimension.activeType = 'relative';
                } else {
                    // set absolute
                    timeDimension.types.absolute.start = (moment($routeParams.start, DATE_TIME_FORMAT))._d;
                    timeDimension.activeType = 'absolute';
                    if(timeObj.end) {
                        matches = timeObj.end.match(/(\d+)([s|m|h|d|w])-ago/);
                        if(matches) return
                        timeDimension.types.absolute.end = (moment($routeParams.end, DATE_TIME_FORMAT))._d;
                    }

                }
            }
        }

        function formattedDateTime(d) {
            return d.getUTCFullYear()+'.'+padTime(d.getMonth()+1)+'.'+padTime(d.getDate())+
                '-'+padTime(d.getHours())+':'+padTime(d.getMinutes())+':'+padTime(d.getSeconds());
        }

        function relativeTimeString(relTimeObj) {
            //return relTimeObj.value.toString()+relTimeObj.unit+'-ago';
            return relTimeObj.start+'-ago';
        }

        function getTimeWindow(td) {
            var tmp = {};
            if(td.activeType === 'relative') {
                tmp.start = relativeTimeString(td.types.relative);
            } else {
                tmp.start = formattedDateTime(td.types.absolute.start);
                //if(td.types.absolute.end !== null) {
                tmp.end = formattedDateTime(td.types.absolute.end);
                //}
            }
            //console.log(tmp);
            return tmp;
        }

        function setTimeWindow(td) {            
            var tmp = $location.search();
            if(td.activeType === 'relative' && tmp.end) delete tmp.end;

            $.extend(true, tmp, getTimeWindow(td), true);
            if(tmp.tags !== undefined && tmp.tags === '') delete tmp.tags;
            $location.search(tmp);
        }

        function _init() {
            if($routeParams.start) parseTime($routeParams);
            scope.timeDimension = timeDimension;

            t.relativeTimeString = relativeTimeString;
            t.setTimeWindow = setTimeWindow;
            t.getTimeWindow = getTimeWindow;
        }

        _init();
    };
    return (TimeWindowManager);
}])
.factory("AnnoFilterManager", ['$routeParams', function($routeParams) {

    var AnnoFilterManager = function(scope) {
        var t = this;

        function parseTypes() {
            var out = [];
            if($routeParams.types && $routeParams.types !== '') {
                var typesArr = $routeParams.types.split(",");
                for(var i=0;i<typesArr.length;i++) {
                    if (typesArr[i] === '') continue
                    out.push(typesArr[i]);
                }
            }
            return out;
        }

        function parseTags() {
            var tags = {};
            if($routeParams.tags) {
                var tagkvs = $routeParams.tags.split(",")
                for(var i=0; i < tagkvs.length; i++) {
                    var kv = tagkvs[i].split(":");
                    if(kv.length != 2) {
                        console.log('invalid tag:' + tagkvs[i]);
                        continue;
                    }
                    tags[kv[0]] = kv[1];
                }
            }
            return tags;
        }

        function removeTag(annoTagsObj, key) {
            if(annoTagsObj[key]) delete annoTagsObj[key];
        }

        function tags2string(obj) {
            var out = '';
            for(var k in obj) {
                out += k+':'+obj[k]+',';
            }
            return out.replace(/\,$/, '');
        }

        function types2string(arr) {
            return arr.join(",");
        }

        function _init() {
            scope.annoFilter = {
                types: parseTypes(),
                tags: parseTags()
            }
            
            t.removeTag = removeTag;
            t.tags2string = tags2string;
            t.types2string = types2string;
        }

        _init();
    };
    return (AnnoFilterManager);
}]);
