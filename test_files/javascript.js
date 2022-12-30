!function ($) {
  $(function () {
    "use strict"; // jshint ;_;
    /* CSS TRANSITION SUPPORT (http://www.modernizr.com/)
    /* ======================================================= */

    $.support.transition = (function () {
      var transitionEnd = (function () {
        var el = document.createElement('bootstrap')
          , transEndEventNames = {
               'WebkitTransition' : 'webkitTransitionEnd'
            }
          , name
        for (name in transEndEventNames){
        }
      }())
    })()
  })

}(window.jQuery);/* ==========================================================
 * bootstrap-alert.js v2.0.3
 * http://twitter.github.com/bootstrap/javascript.html#alerts
 * ==========================================================
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * ========================================================== */
// Blank = 2, Comment = 9,  Code = 16, Total = 27
