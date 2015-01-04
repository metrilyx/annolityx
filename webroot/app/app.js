/* helper functions */
function padTime(val) {
	if(val < 10) return "0"+val.toString();
	return val.toString();
}

var app = angular.module('app', [
	'ngRoute',
	'ngResource',
	'ui.bootstrap.datetimepicker',
	'app.directives',
	'app.factories',
	'app.controllers',
	'app.services'
]);

(function() {
	/*
	 * Bootstrap the app with the config fetched via http
	 */
	var configConstant = "AnnolityxConfig";
	var configUrl      = "/api/config";

    function fetchAndInjectConfig() {
        var initInjector = angular.injector(["ng"]);
        var $http = initInjector.get("$http");

        return $http.get(configUrl).then(function(response) {
            app.constant(configConstant, response.data);
        }, function(errorResponse) {
            // Handle error case
            console.log(errorResponse);
        });
    }

    function bootstrapApplication() {
        angular.element(document).ready(function() {
            angular.bootstrap(document, ["app"]);
        });
    }

    fetchAndInjectConfig().then(bootstrapApplication);
    
}());

app.config(['$routeProvider',
	function($routeProvider) {
		$routeProvider.
			when('/login', {
				templateUrl: 'partials/login.html',
				controller: 'loginController'
			}).
			when('/', {
				templateUrl: 'partials/root.html',
				controller: 'rootController'
			}).
			otherwise({
				redirectTo: '/login'
			});
	}
]);

app.filter('objectLength', function() {
	return function(obj) {
    	return Object.keys(obj).length;
	};
}).filter('datetimeFromEpoch', function() {
	//var days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	//var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
	
	return function(epoch) {
		return (new Date(epoch*1000)).toString()
		//return d.toString();
		//return padTime(d.getHours())+':'+padTime(d.getMinutes())+':'+padTime(d.getSeconds())+' '+
		//	days[d.getDay()]+' '+months[d.getMonth()]+' '+padTime(d.getDate())+', '+d.getUTCFullYear();
	}
});
