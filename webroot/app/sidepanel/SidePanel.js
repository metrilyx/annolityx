angular.module('annolityx.sidepanel', [])
.directive('sidepanel', [function() {
    return {
        restrict: 'E',
        templateUrl: 'app/sidepanel/sidepanel.html',
        link: function(scope, elem, attrs) {

        }
    };
}])
.directive('annotationTypes', ['$rootScope', '$location', 'EventAnnotationTypes', 
    function($rootScope, $location, EventAnnotationTypes) {
    
    return {
        restrict: 'EA',
        templateUrl: 'app/sidepanel/anno-types.html',
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
                    scope.annoTypesIndex = indexAnnoTypesList(data);
                
                    /* Watch for filter changes */
                    scope.$watch(function() { return scope.annoTypesIndex; }, onAnnoFilterChange, true);
                
                }, function(err){
                    console.log(err);
                });
            }

            init();
        }
    };
}]);