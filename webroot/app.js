/* helper functions */
function padTime(val) {
	if(val < 10) return "0"+val.toString();
	return val.toString();
}

var app = angular.module('app', [
	'ngRoute',
	'ngResource',
	'appDirectives',
	'appFactories',
	'appControllers',
	'ui.bootstrap.datetimepicker',
	'appServices'
]);

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
	var days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
	var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
	
	return function(epoch) {
		return (new Date(epoch*1000)).toString()
		//return d.toString();
		//return padTime(d.getHours())+':'+padTime(d.getMinutes())+':'+padTime(d.getSeconds())+' '+
		//	days[d.getDay()]+' '+months[d.getMonth()]+' '+padTime(d.getDate())+', '+d.getUTCFullYear();
	}
});
