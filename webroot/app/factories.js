'use strict';

angular.module("app.factories", [])
.factory("WebSocketManager", [ 'AnnolityxConfig', function(AnnolityxConfig) {
    /* this would be persistent across controller navigation */
    var _wsock = null;

    var WebSocketManager = function(cb, messageFilter) {
        //console.log('WebSocketManager');
        var t = this;

        var msgCallback = cb,
            messageFilter = messageFilter,
            maxRetries = 3,
            retryCount = 0,
            retryInterval = 10000;
        

        var sendMessage = function(data) {
            
            _wsock.send(angular.toJson(data));
        }

        var onWsOpen = function(evt) {
            
            console.log("Connection open. Extentions: [ '" + _wsock.extensions + "' ]");
            retryCount = 0;
            sendMessage(messageFilter);
        }

        var onWsClose = function(evt) {
            
            console.log('Connection closed: code='+evt.code+
                '; reason='+evt.reason+
                '; wasClean='+evt.wasClean);
            
            _wsock = null;
            
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

        var msgErrback = function(evt) {
            console.error('Subscriber error:', evt.data);
            console.warn(evt);
        }

       var onWsMessage = function(evt) {
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

        var connect = function() {
            _wsock = new WebSocket(AnnolityxConfig.websocket.url);
            _wsock.addEventListener('open', onWsOpen);
            _wsock.addEventListener('message', onWsMessage);
            _wsock.addEventListener('close', onWsClose);
        }

        var disconnect = function(reason) {
            
            if ( _wsock && _wsock.readyState == 1 ) {
                //console.log('Already connected!');
                _wsock.removeEventListener('close', onWsClose);
                // 1000 = normal close
                _wsock.close(1000, reason);    
            }
        }

        function _initialize() {
            
            for(var k in messageFilter) {
                /* delete invalid keys */
                if(k !== 'tags' && k !== 'types') delete messageFilter[k];
            }
            
            t.sendMessage = sendMessage
            t.connect = connect;
            t.disconnect = disconnect;
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
