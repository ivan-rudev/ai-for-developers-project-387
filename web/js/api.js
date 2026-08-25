/* ==========================================================================
   API-клиент — контракт api/generated/openapi.yaml (REST, JSON).
   При ошибке бросает ApiError со статусом и сообщением сервера.
   ========================================================================== */

(function (global) {
  'use strict';

  class ApiError extends Error {
    constructor(status, code, message) {
      super(message || 'Неизвестная ошибка');
      this.name = 'ApiError';
      this.status = status;
      this.code = code;
    }
  }

  async function request(path, options) {
    const opts = options || {};
    const res = await fetch(path, {
      method: opts.method || 'GET',
      headers: { 'Content-Type': 'application/json' },
      body: opts.body,
    });

    const text = await res.text();
    let body = null;
    if (text) {
      try {
        body = JSON.parse(text);
      } catch (e) {
        body = { raw: text };
      }
    }

    if (!res.ok) {
      const code = body && body.error ? body.error : 'request_failed';
      const message = body && body.message ? body.message : 'Ошибка запроса к серверу';
      throw new ApiError(res.status, code, message);
    }

    return body;
  }

  const API = {
    getOwners: function () {
      return request('/api/owners');
    },

    getOwner: function (uuid) {
      return request('/api/owners/' + encodeURIComponent(uuid));
    },

    createOwner: function (data) {
      return request('/api/owners', { method: 'POST', body: JSON.stringify(data) });
    },

    getOwnerEvents: function (uuid) {
      return request('/api/owners/' + encodeURIComponent(uuid) + '/events');
    },

    getOwnerBookings: function (uuid) {
      return request('/api/owners/' + encodeURIComponent(uuid) + '/bookings');
    },

    getSlots: function (uuid, eventUuid) {
      return request('/api/owners/' + encodeURIComponent(uuid) + '/slots?event_uuid=' + encodeURIComponent(eventUuid));
    },

    createBooking: function (data) {
      return request('/api/bookings', { method: 'POST', body: JSON.stringify(data) });
    },

    getAdmin: function () {
      return request('/api/admin');
    },

    getAdminBookings: function () {
      return request('/api/admin/bookings');
    },

    getAdminEvents: function () {
      return request('/api/admin/events');
    },

    createAdminEvent: function (data) {
      return request('/api/admin/events', { method: 'POST', body: JSON.stringify(data) });
    },
  };

  global.API = API;
  global.ApiError = ApiError;
})(window);
