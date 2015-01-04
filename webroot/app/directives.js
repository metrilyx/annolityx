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
.directive('annotationTypes', ['$location', function($location) {
	return {
		restrict: 'A',
		require: '?ngModel',
		link: function(scope, elem, attrs, ctrl) {
			if(!ctrl) return;
		
			scope.$watch(function(){return ctrl.$modelValue;}, function(newVal, oldVal) {
				
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
				if(tmp.tags !== undefined && tmp['tags'] == "") {
					delete tmp['tags'];
				}

				tmp.types = newtypes.replace(/\,$/, '');
				if(tmp.types != '') $location.search(tmp);
			}, true);
		}
	};
}]);