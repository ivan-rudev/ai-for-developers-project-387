/* ==========================================================================
   Простой клиентский роутер на history.pushState.
   Регистрирует маршруты по паттернам вида "/owners/:uuid" и вызывает
   обработчики при переходе по SPA-адресам.
   ========================================================================== */

(function (global) {
  'use strict';

  var routes = [];
  var fallback = function () {
    global.location.href = '/';
  };

  function escapeRegExp(str) {
    return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  function compile(pattern) {
    var names = [];
    var norm = '/' + String(pattern || '/').replace(/^\/+|\/+$/g, '');
    var parts = norm
      .split('/')
      .filter(Boolean)
      .map(function (part) {
        if (part.charAt(0) === ':') {
          names.push(part.slice(1));
          return '([^/]+)';
        }
        return escapeRegExp(part);
      });
    var body = parts.length ? '/' + parts.join('/') : '';
    var source = '^' + body + '/?$';
    return { re: new RegExp(source), names: names };
  }

  function resolve(path) {
    var clean = String(path || '/').split('?')[0];
    for (var i = 0; i < routes.length; i++) {
      var route = routes[i];
      var match = clean.match(route.re);
      if (match) {
        var params = {};
        for (var j = 0; j < route.names.length; j++) {
          try {
            params[route.names[j]] = decodeURIComponent(match[j + 1]);
          } catch (e) {
            params[route.names[j]] = match[j + 1];
          }
        }
        return { handler: route.handler, params: params };
      }
    }
    return null;
  }

  function dispatch() {
    var match = resolve(global.location.pathname);
    if (match) {
      match.handler(match.params);
    } else {
      fallback();
    }
  }

  function register(pattern, handler) {
    var compiled = compile(pattern);
    routes.push({
      pattern: pattern,
      re: compiled.re,
      names: compiled.names,
      handler: handler,
    });
  }

  function navigate(path) {
    global.history.pushState({}, '', path);
    dispatch();
  }

  function replace(path) {
    global.history.replaceState({}, '', path);
    dispatch();
  }

  global.addEventListener('popstate', dispatch);
  document.addEventListener('DOMContentLoaded', dispatch);

  global.Router = {
    register: register,
    navigate: navigate,
    replace: replace,
    dispatch: dispatch,
  };
})(window);
