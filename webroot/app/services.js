angular.module('app.services', [])
.factory('Authenticator', ['$window', '$http', '$location', '$routeParams', function($window, $http, $location, $routeParams) {

	function _sessionIsAuthenticated() {
		if($window.sessionStorage['credentials']) {

			var creds = JSON.parse($window.sessionStorage['credentials']);
			if(creds.username && creds.username !== "" && creds.password && creds.password !== "") {
				// do custom checking here
				return true
			}
		}
		return false;
	}

	function _login(creds) {
		// do actual auth here //
		if(creds.username === "guest" && creds.password === "guest") {
			$window.sessionStorage['credentials'] = JSON.stringify(creds);
			return true;
		}
		return false;
	}

	function _logout() {
		var sStor = $window.sessionStorage;
		if(sStor['credentials']) {
			delete sStor['credentials'];
		}
		$location.url("/login");
	}

    function _checkAuthOrRedirect() {
        if(!_sessionIsAuthenticated()) {
            $location.url("/login?redirect="+$location.url());
            return false;
        } else {
            return true;
        }
    }

	var Authenticator = {
        login                 : _login,
        logout                : _logout,
		sessionIsAuthenticated: _sessionIsAuthenticated,
		checkAuthOrRedirect   : _checkAuthOrRedirect
    };

    return (Authenticator);
}])
.factory('EvtAnnoService', ['$resource', function($resource) {
    return $resource('/api/annotations', {}, {
        search: {
            method: 'GET',
            isArray: true
        }
    });
}])
.factory('EventAnnotationTypes', ['$http', function($http) {
    var _cache = null;
    
    return {
        list: function() {
            var _d = $.Deferred();
            if (_cache != null && Object.keys(_cache).length > 0) {
                _d.resolve(_cache);
            } else {
                $http({
                    method: 'GET',
                    url: '/api/types',
                }).success(function(data) {
                    _cache = data;
                    _d.resolve(_cache);
                }).error(function(err) {
                    _d.reject(err);
                });
            }

            return _d;
        }
    };
}]);




