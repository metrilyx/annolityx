angular.module('app.services', [])
.factory('Authenticator', ['$window', '$http', '$location', function($window, $http, $location) {

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

    function _checkAuthOrRedirect(redirectTo) {
        if(!_sessionIsAuthenticated()) $location.url("/login?redirect="+redirectTo);
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
.factory('EventAnnotationTypes', ['$resource', function($resource) {
    /* TODO: make this a cached call */
    return $resource('/api/types/:annoType', {}, {
        list: {
            method: 'GET',
            isArray: true,
        }
    });
}]);




