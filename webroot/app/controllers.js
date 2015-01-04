angular.module('app.controllers', [])
.controller('rootController', [
	'$window', '$scope', '$location', 'Authenticator',
	'EvtAnnoService', 'EventAnnotationTypes', 'WebSocketManager', 'TimeWindowManager', 'AnnoFilterManager',
	function($window, $scope, $location, Authenticator, 
		EvtAnnoService, EventAnnotationTypes, WebSocketManager, TimeWindowManager, AnnoFilterManager) {

		Authenticator.checkAuthOrRedirect("/");

		$scope.pageHeaderHtml    = "/partials/page-header.html";
		$scope.sidePanelHtml     = "/partials/side-panel.html";
		$scope.timeSelectorHtml  = "/partials/time-selector.html";

		var annoFilterMgr = new AnnoFilterManager($scope);
		var timeWinMgr = new TimeWindowManager($scope);
		var webSockMgr;

		$scope.tagKeyValueInput = "";

		$scope.annoTypesIndex = {};
		
		$scope.annoResults = [];
		$scope.maxResults = 100;
		$scope.annoSortKey = "timestamp";
		$scope.annoSortReverse = true;

		$scope.annoMessageSearch = "";

		function sortAnnoByKey(sortKey) {
			if($scope.annoSortKey === sortKey) {
				$scope.annoSortReverse = !$scope.annoSortReverse;
			} else {
				$scope.annoSortKey = sortKey;
				if(sortKey === 'timestamp') $scope.annoSortReverse = true;
				else $scope.annoSortReverse = false;
			}
		}

		function removeAnnoTag(tagKey) {
			annoFilterMgr.removeTag($scope.annoFilter.tags, tagKey);

			var tmp = $location.search();
			tmp.tags = annoFilterMgr.tags2string($scope.annoFilter.tags);
			if(tmp.tags === '') delete tmp.tags;

			$location.search(tmp);
		}

		function onWebsockData(wsData) {
			$scope.$apply(function() {
				$scope.annoResults.unshift(wsData);
			});
		}

		function indexAnnoTypesList(data) {
			var o = {};
			if($scope.annoFilter.types.length < 1) {
				for(var i=0; i < data.length; i++) {
					data[i].selected = true;
					o[data[i].id] = data[i];
				}
			} else {
				for(var i=0; i < data.length; i++) {
					data[i].selected = false;
					for(var t=0; t<$scope.annoFilter.types.length; t++) {
						
						if($scope.annoFilter.types[t] == data[i].id) 
							data[i].selected = true;
					}
					o[data[i].id] = data[i];
				}
			}
			return o;
		}

		function setTimeRange() {
			timeWinMgr.setTimeWindow($scope.timeDimension);
		}

		function getAnnoQuery() {
			var query = $.extend({}, $scope.annoFilter, timeWinMgr.getTimeWindow($scope.timeDimension), true);
		
			query.tags = annoFilterMgr.tags2string($scope.annoFilter.tags);
			if(query.tags === '') delete query.tags;

			query.types = annoFilterMgr.types2string($scope.annoFilter.types);
			if(query.types === '') delete query.types;

			return query;
		}

		function _initialize() {
			
			$scope.setTimeRange = setTimeRange;
			$scope.sortAnnoByKey = sortAnnoByKey;
			$scope.removeAnnoTag = removeAnnoTag;

			/* TODO: only if relative time */
			if($scope.timeDimension.activeType == 'relative') {
				console.log("Live events enabled.");
				webSockMgr = new WebSocketManager(onWebsockData, $scope.annoFilter);
				webSockMgr.connect();
			}

			EventAnnotationTypes.list(function(data) {
				$scope.annoTypesIndex = indexAnnoTypesList(data);
			});

			EvtAnnoService.search(getAnnoQuery(), function(data) {
				$scope.annoResults = data;
				console.log(data.length);
			});
		}

		_initialize();
	}
])
.controller('defaultController', [ '$window', '$location', '$scope',
	function($window, $location, $scope) {

		$scope.pageHeaderHtml = "/partials/page-header.html";

		$scope.logOut = function() {

	        console.log("De-authing...");
	        var sStor = $window.sessionStorage;
	        if(sStor['credentials']) {
	            delete sStor['credentials'];
	        }
	        /*
	        var lStor = $window.localStorage;
	        for(var k in lStor) {
	            if(/^token\-/.test(k)) delete lStor[k];
	        }
			*/
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
