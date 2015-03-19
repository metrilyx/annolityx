angular.module('app.directives', [])
.directive('appStage', ['$window', function($window) {
	return {
		restrict: 'A',
		require: '?ngModel',
		link: function(scope, elem, attrs, ctrl) {
			if(!ctrl) return;

			var stage = { jDom: $(elem) };
			stage.fillScreen = function() {

				if(stage.jDom) {

					stage.jDom.css("width", $window.innerWidth - stage.jDom.scrollWidth);
					stage.jDom.css("height", $window.innerHeight);
				}
			};

			function init() {

				stage.fillScreen();
				$window.addEventListener("resize", function(event) { $stage.fillScreen() });
			}

			init();
		}
	}
}])
.directive('tagKeyValue', ['$location', function($location) {
	return {
		restrict: 'A',
		require: '?ngModel',
		link: function(scope, elem, attrs, ctrl) {
			if(!ctrl) return;

			function tags2string(obj) {
				var out = '';
				for(var k in obj) {
					out += k+':'+obj[k]+',';
				}
				return out.replace(/\,$/, '');
			}

			elem[0].addEventListener('keyup', function(evt) {
				if(evt.keyCode == 13) {

					var kv = ctrl.$viewValue.split("=");
					if(kv.length == 2) {
						if(kv[1] !== undefined && kv[1] !== ''){
							ctrl.$setValidity('tagkeyvalue', true);
							scope.$apply(function() { ctrl.$modelValue[kv[0]] = kv[1]; });
							elem[0].value = '';
							return;
						}
					}
				}
				ctrl.$setValidity('tagkeyvalue', false);
			});

			// model 2 view
			ctrl.$formatters.push(function(modelValue) { return ""; });
			// view 2 model
			ctrl.$parsers.unshift(function(viewValue) { return ctrl.$modelValue; });
			
			scope.$watch(function() { return ctrl.$modelValue; }, 
				function(newVal, oldVal) {
				
					var tmp = $location.search();
					tmp.tags = tags2string(newVal);
					if(tmp.tags != '') $location.search(tmp);
				}, true);
		}
	};
}])
.directive('annotationTypes', ['$rootScope', '$location', 'EventAnnotationTypes', 
	function($rootScope, $location, EventAnnotationTypes) {
	
	return {
		restrict: 'EA',
		templateUrl: 'partials/anno-types.html',
		link: function(scope, elem, attrs) {

			var indexAnnoTypesList = function(data) {
				var o = {};
				if(scope.annoFilter.types.length < 1) {
					for(var i=0; i < data.length; i++) {
						data[i].selected = true;
						o[data[i].id] = data[i];
					}
				} else {
					for(var i=0; i < data.length; i++) {
						data[i].selected = false;
						for(var t=0; t < scope.annoFilter.types.length; t++) {
							
							if(scope.annoFilter.types[t] == data[i].id) 
								data[i].selected = true;
						}
						o[data[i].id] = data[i];
					}
				}
				return o;
			}

			var onAnnoFilterChange = function(newVal, oldVal) {
				if(!newVal) return;

				/* Happens at times during initialization */
				if ( angular.equals(newVal, oldVal) ) return;

				if(Object.keys(oldVal).length==0 && Object.keys(newVal).length==0) {
					return;
				} else if(Object.keys(oldVal).length == 0 && Object.keys(newVal).length > 0) {
					return;
				}

				var newtypes = '';
				for(var k in newVal) {
					if (newVal[k].selected) newtypes += newVal[k].id+',';
				}

				var tmp = $location.search();
				if(tmp.tags !== undefined && tmp['tags'] == "") delete tmp['tags'];
				
				tmp.types = newtypes.replace(/\,$/, '');
				if(tmp.types != '') $location.search(tmp);
			}

			var init = function() {
				EventAnnotationTypes.list().then(function(data) {
					// Apply to parent scope or changes will not reflect.
					scope.$parent.annoTypesIndex = indexAnnoTypesList(data);
				
					/* Watch for filter changes */
					scope.$watch(function() { return scope.annoTypesIndex; }, onAnnoFilterChange, true);
				
				}, function(err){
					console.log(err);
				});
			}

			init();
		}
	};
}])
.directive('annotationDetail', [function() {
	return {
		restrict: 'A',
		templateUrl: 'partials/anno-details.html',
		link: function(scope, elem, attrs, ctrl) {}
	};
}]);