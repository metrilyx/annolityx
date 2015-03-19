angular.module('app.controllers', [])
.controller('rootController', [
	'$rootScope', '$window', '$scope', '$location', 'Authenticator',
	'EvtAnnoService', 'WebSocketManager', 'TimeWindowManager', 'AnnoFilterManager',
	function($rootScope, $window, $scope, $location, Authenticator, 
		EvtAnnoService, WebSocketManager, TimeWindowManager, AnnoFilterManager) {

		if(!Authenticator.checkAuthOrRedirect()) return;

		console.log('called');

		$scope.pageHeaderHtml    = "partials/page-header.html";
		$scope.sidePanelHtml     = "partials/side-panel.html";
		$scope.timeSelectorHtml  = "partials/time-selector.html";

		$scope.tagKeyValueInput = "";
		
		$scope.annoTypesIndex = {};

		$scope.annoResults = [];
		$scope.maxResults = 100;
		$scope.annoSortKey = "timestamp";
		$scope.annoSortReverse = true;

		$scope.annoMessageSearch = "";

		var annoFilterMgr = new AnnoFilterManager($scope),
			timeWinMgr 	  = new TimeWindowManager($scope),
			webSockMgr    = new WebSocketManager(onWebsockData, $scope.annoFilter);
		
		var onWebsockData = function(wsData) {
			$scope.$apply(function() {
				$scope.annoResults.unshift(wsData);
			});
		}

		var getAnnoQuery = function() {
			var query = $.extend({}, $scope.annoFilter, timeWinMgr.getTimeWindow($scope.timeDimension), true);
		
			query.tags = annoFilterMgr.tags2string($scope.annoFilter.tags);
			if(query.tags === '') delete query.tags;

			query.types = annoFilterMgr.types2string($scope.annoFilter.types);
			if(query.types === '') delete query.types;

			return query;
		}

		$scope.sortAnnoByKey = function(sortKey) {
			if($scope.annoSortKey === sortKey) {
				$scope.annoSortReverse = !$scope.annoSortReverse;
			} else {
				$scope.annoSortKey = sortKey;
				if(sortKey === 'timestamp') $scope.annoSortReverse = true;
				else $scope.annoSortReverse = false;
			}
		}

		$scope.removeAnnoTag = function(tagKey) {
			annoFilterMgr.removeTag($scope.annoFilter.tags, tagKey);

			var tmp = $location.search();
			tmp.tags = annoFilterMgr.tags2string($scope.annoFilter.tags);
			if(tmp.tags === '') delete tmp.tags;

			$location.search(tmp);
		}
		
		$scope.setTimeRange = function() {
			timeWinMgr.setTimeWindow($scope.timeDimension);
		}

		$scope.toggleAnnoDetails = function(_id) {
			
			$("[data-anno-id='"+_id+"']").toggle();
		}

		function _initialize() {

			if($scope.timeDimension.activeType == 'relative') {
				console.log("Live events enabled.");
				webSockMgr.connect();
			}

			EvtAnnoService.search(getAnnoQuery(), function(data) {
				$scope.annoResults = data;
				console.log("Results:", data.length);
			});

			$scope.$on('$destroy', function() { 
				webSockMgr.disconnect('Controller destroyed'); });
		}
		
		_initialize();
	}
])
.controller('defaultController', [ '$window', '$location', '$scope',
	function($window, $location, $scope) {

		$scope.logOut = function() {

	        console.log("De-authing...");
	        var sStor = $window.sessionStorage;
	        if(sStor['credentials']) {
	            delete sStor['credentials'];
	        }
	        //var lStor = $window.localStorage;
	        $location.url("/login");
	    }
	}
])
.controller('loginController', [
	'$scope', '$window', '$routeParams', '$location', 'Authenticator',
	function($scope, $window, $routeParams, $location, Authenticator) {

		var defaultPage = "/";

		$scope.credentials = { username: "guest", password: "guest" };

		$scope.attemptLogin = function() {
			if(Authenticator.login($scope.credentials)) {

				if($routeParams.redirect) $location.url($routeParams.redirect);
				else $location.url(defaultPage);
			} else {

				$("#login-window-header").html("<span>Auth failed!</span>");
			}
		}

		function _initialize() {
			if($window.sessionStorage['credentials']) {

				var creds = JSON.parse($window.sessionStorage['credentials']);
				if(creds.username && creds.username !== "" && creds.password && creds.password !== "") {

					$scope.credentials = creds;
					$scope.attemptLogin();
				}
			}
		}

		_initialize();
	}
]);
