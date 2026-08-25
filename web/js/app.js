/* ==========================================================================
   Логика SPA — экраны по docs/PRD.md §9 и дизайн-система docs/DESIGN.md.
   Тёмная тема (theme-dark) — личный кабинет "/office", светлая — "/",
   /owners и визард.
   ========================================================================== */

(function () {
  'use strict';

  var app = document.getElementById('app');

  var state = {
    adminTab: 'bookings',
    adminOwner: null,
    adminBookings: [],
    adminEvents: [],
    owner: null,
    events: [],
    bookings: [],
    event: null,
    slots: null,
    selectedDate: null,
    selectedSlot: null,
    booking: null,
    guestName: '',
    guestEmail: '',
  };

  /* ================================ Утилиты ================================ */

  function esc(str) {
    return String(str == null ? '' : str)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  function render(html) {
    app.innerHTML = html;
  }

  function loadingHtml() {
    return '<div class="loading">Загрузка…</div>';
  }

  function errorHtml(err) {
    return '<div class="page"><div class="empty">' + esc(err && err.message ? err.message : 'Произошла ошибка') + '</div></div>';
  }

  function setTheme(dark) {
    document.body.classList.toggle('theme-dark', dark);
  }

  function pad(n) {
    return String(n).padStart(2, '0');
  }

  function parseDate(str) {
    var p = String(str).split('-');
    return { y: +p[0], m: +p[1], d: +p[2] };
  }

  function dateKey(d) {
    return d.y + '-' + pad(d.m) + '-' + pad(d.d);
  }

  function slotEnd(time, durationMinutes) {
    var p = time.split(':');
    var total = +p[0] * 60 + +p[1] + durationMinutes;
    return pad(Math.floor(total / 60)) + ':' + pad(total % 60);
  }

  var WEEKDAY_SHORT = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];
  var MONTH_GEN = ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'];

  function formatDateShort(dateStr) {
    var d = parseDate(dateStr);
    return d.d + ' ' + MONTH_GEN[d.m - 1];
  }

  function formatMonthYear(dateStr) {
    var d = parseDate(dateStr);
    return new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric', timeZone: 'UTC' })
      .format(new Date(Date.UTC(d.y, d.m - 1, 1)));
  }

  function formatInTz(iso, tz, opts) {
    return new Intl.DateTimeFormat('ru-RU', Object.assign({ timeZone: tz }, opts)).format(new Date(iso));
  }

  function timeInTz(iso, tz) {
    return formatInTz(iso, tz, { hour: '2-digit', minute: '2-digit' });
  }

  function dateShortInTz(iso, tz) {
    return formatInTz(iso, tz, { day: 'numeric', month: 'short' });
  }

  function hasAvailable(daySlots) {
    return daySlots.some(function (s) { return s.status === 'available'; });
  }

  function plural(n, one, few, many) {
    var m10 = n % 10;
    var m100 = n % 100;
    if (m10 === 1 && m100 !== 11) return one;
    if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) return few;
    return many;
  }

  function errorText(err, map) {
    if (err instanceof ApiError) {
      if (map && map[err.code]) return map[err.code];
      if (err.code === 'conflict') return 'Такой объект уже существует';
      if (err.code === 'rate_limit') return 'Слишком много запросов. Попробуйте позже.';
      if (err.code === 'invalid_input') return 'Проверьте правильность заполнения полей';
      return err.message;
    }
    return err && err.message ? err.message : 'Неизвестная ошибка';
  }

  /* ============================== Шаблоны =============================== */

  function topbarTemplate() {
    return (
      '<header class="topbar">' +
        '<a class="topbar__brand" href="/" data-nav="/">Calendar MVP</a>' +
        '<div class="topbar__actions">' +
          '<button class="btn btn--secondary" data-nav="/owners/new">Создать свой календарь</button>' +
        '</div>' +
      '</header>'
    );
  }

  function stepsHtml(current) {
    var items = [];
    for (var i = 1; i <= 5; i++) {
      var cls = i === current ? ' steps__item--current' : (i < current ? ' steps__item--done' : '');
      items.push(
        '<div class="steps__item' + cls + '">' +
          (i > 1 ? '<span class="steps__line"></span>' : '') +
          '<span class="steps__dot">' + i + '</span>' +
        '</div>'
      );
    }
    return '<div class="steps">' + items.join('') + '</div>';
  }

  function eventCardHtml(e) {
    return (
      '<div class="card card--event" data-action="select-event" data-event-uuid="' + esc(e.uuid) + '" role="button" tabindex="0">' +
        '<span class="card--event__name">' + esc(e.name) + '</span>' +
        '<span class="badge card--event__badge">' + e.duration_minutes + 'м</span>' +
        '<span class="card--event__desc">' + esc(e.description || '') + '</span>' +
      '</div>'
    );
  }

  /* ======================= Приветственная страница ("/") =================== */

  function renderLanding() {
    setTheme(false);
    document.title = 'Calendar MVP — запись на встречи';
    render(
      topbarTemplate() +
      '<div class="page page--narrow">' +
        '<div class="wizard-head">' +
          '<h1>Calendar MVP — запись на встречи</h1>' +
          '<p class="subtitle">Сервис бронирования временных слотов: выбирайте владельца календаря, тип встречи и свободное время — и записывайтесь за пару кликов.</p>' +
        '</div>' +
        '<section class="section">' +
          '<h2 class="section__title">О проекте</h2>' +
          '<p class="muted">Calendar MVP — учебный сервис записи на звонки (упрощённый аналог Cal.com). Владельцы календарей создают публичные страницы с настройками доступности, а гости выбирают тип встречи (15 или 30 минут) и бронируют свободный слот. Всё без регистрации: достаточно имени и email.</p>' +
        '</section>' +
        '<section class="section">' +
          '<h2 class="section__title">Начать</h2>' +
          '<div class="card-grid">' +
            '<div class="card card--owner" data-nav="/owners" role="link" tabindex="0">' +
              '<span class="card--owner__name">Доступные календари</span>' +
              '<span class="btn btn--primary btn--sm">Посмотреть →</span>' +
            '</div>' +
            '<div class="card card--owner" data-nav="/office" role="link" tabindex="0">' +
              '<span class="card--owner__name">Личный кабинет владельца</span>' +
              '<span class="btn btn--primary btn--sm">Открыть →</span>' +
            '</div>' +
          '</div>' +
          '<p class="muted text-sm">Демо-проект без аутентификации: личный кабинет открыт и привязан к владельцу из config.yaml. Не использовать в публичной сети и с реальными данными.</p>' +
        '</section>' +
      '</div>'
    );
  }

  /* ==================== Личный кабинет владельца ("/office") =============== */

  function renderAdmin() {
    setTheme(true);
    document.title = 'Calendar MVP Admin';
    render(loadingHtml());
    Promise.all([API.getAdmin(), API.getAdminBookings(), API.getAdminEvents()])
      .then(function (r) {
        state.adminOwner = r[0];
        state.adminBookings = r[1].bookings;
        state.adminEvents = r[2].events;
        renderAdminPage();
      })
      .catch(function (err) {
        render(errorHtml(err));
      });
  }

  function renderAdminPage() {
    var owner = state.adminOwner;
    render(
      '<header class="topbar">' +
        '<a class="topbar__brand" href="/" data-nav="/">Calendar MVP <span class="topbar__brand-sub">Admin</span></a>' +
        '<span class="muted text-sm">' + esc(owner.name) + ' · ' + esc(owner.settings.timezone) + ', ' + esc(owner.settings.work_start) + '–' + esc(owner.settings.work_end) + '</span>' +
      '</header>' +
      '<div class="page page--admin">' +
        '<div class="tabs">' +
          '<button class="tab' + (state.adminTab === 'bookings' ? ' tab--active' : '') + '" data-tab="bookings">Бронирования</button>' +
          '<button class="tab' + (state.adminTab === 'events' ? ' tab--active' : '') + '" data-tab="events">Типы встреч</button>' +
        '</div>' +
        '<div class="admin-body">' +
          (state.adminTab === 'bookings' ? adminBookingsHtml() : adminEventsHtml()) +
        '</div>' +
      '</div>'
    );
  }

  function adminBookingsHtml() {
    if (!state.adminBookings.length) {
      return '<div class="empty">Нет предстоящих бронирований</div>';
    }
    var tz = state.adminOwner.settings.timezone;
    var rows = state.adminBookings.map(function (b) {
      return (
        '<tr>' +
          '<td>' + esc(dateShortInTz(b.start_time, tz)) + ', ' + esc(timeInTz(b.start_time, tz)) + '–' + esc(timeInTz(b.end_time, tz)) + '</td>' +
          '<td>' + esc(b.event_name) + '</td>' +
          '<td>' + esc(b.guest_name) + '</td>' +
          '<td>' + esc(b.guest_email) + '</td>' +
        '</tr>'
      );
    }).join('');
    return (
      '<div class="table-wrap">' +
        '<table class="table">' +
          '<thead><tr><th>Дата и время</th><th>Событие</th><th>Гость</th><th>Email</th></tr></thead>' +
          '<tbody>' + rows + '</tbody>' +
        '</table>' +
      '</div>'
    );
  }

  function adminEventsHtml() {
    var rows = state.adminEvents.map(function (e) {
      return (
        '<tr>' +
          '<td class="uuid-cell" title="' + esc(e.uuid) + '">' + esc(e.uuid) + '</td>' +
          '<td>' + esc(e.name) + '</td>' +
          '<td>' + e.duration_minutes + ' мин</td>' +
          '<td>' + (e.is_active
            ? '<span class="badge badge--success">Активно</span>'
            : '<span class="badge badge--muted">Неактивно</span>') + '</td>' +
        '</tr>'
      );
    }).join('');
    return (
      '<div class="table-wrap">' +
        '<table class="table">' +
          '<thead><tr><th>ID (UUID)</th><th>Название</th><th>Длительность</th><th>Активность</th></tr></thead>' +
          '<tbody>' + rows + '</tbody>' +
        '</table>' +
      '</div>' +
      '<form id="admin-event-form" class="section">' +
        '<h3 class="section__title">Добавить тип встречи</h3>' +
        '<div class="form-row">' +
          '<div class="field">' +
            '<label class="field__label" for="event-name">Название <span class="required">*</span></label>' +
            '<input class="input" id="event-name" name="name" type="text" required maxlength="120" placeholder="Например, Консультация" />' +
          '</div>' +
          '<div class="field">' +
            '<label class="field__label" for="event-duration">Длительность (мин) <span class="required">*</span></label>' +
            '<input class="input" id="event-duration" name="duration_minutes" type="number" min="1" max="720" required placeholder="15, 30, 45…" />' +
          '</div>' +
        '</div>' +
        '<div class="field">' +
          '<label class="field__label" for="event-desc">Описание</label>' +
          '<input class="input" id="event-desc" name="description" type="text" placeholder="Короткое описание встречи" />' +
        '</div>' +
        '<div class="form-error" id="admin-event-error"></div>' +
        '<button class="btn btn--primary" type="submit">Добавить</button>' +
      '</form>'
    );
  }

  function submitAdminEvent(form) {
    var name = form.name.value.trim();
    var duration = parseInt(form.duration_minutes.value, 10);
    var description = (form.description.value || '').trim();
    var errorEl = document.getElementById('admin-event-error');
    errorEl.textContent = '';
    API.createAdminEvent({ name: name, description: description, duration_minutes: duration })
      .then(function () {
        form.reset();
        return API.getAdminEvents();
      })
      .then(function (res) {
        state.adminEvents = res.events;
        renderAdminPage();
      })
      .catch(function (err) {
        errorEl.textContent = errorText(err, {
          conflict: 'Событие с таким названием уже существует',
        });
      });
  }

  /* ======================= Список владельцев ("/owners") ===================== */

  function renderOwnersList() {
    setTheme(false);
    document.title = 'Calendar MVP — доступные календари';
    render(topbarTemplate() + loadingHtml());
    API.getOwners()
      .then(function (owners) {
        var cards = owners.map(function (o) {
          return (
            '<div class="card card--owner" data-nav="/owners/' + esc(o.uuid) + '" role="link" tabindex="0">' +
              '<span class="card--owner__name">' + esc(o.name) + '</span>' +
              '<span class="btn btn--primary btn--sm">Записаться на встречу →</span>' +
            '</div>'
          );
        }).join('');
        render(
          topbarTemplate() +
          '<div class="page">' +
            '<h1>Доступные календари</h1>' +
            '<p class="subtitle">Выберите владельца календаря, чтобы записаться на встречу.</p>' +
            '<div class="section">' +
              (owners.length
                ? '<div class="card-grid">' + cards + '</div>'
                : '<div class="empty">Пока нет доступных календарей</div>') +
            '</div>' +
          '</div>'
        );
      })
      .catch(function (err) {
        render(topbarTemplate() + errorHtml(err));
      });
  }

  /* ===================== Создание владельца ("/owners/new") ================= */

  function renderCreateOwner() {
    setTheme(false);
    document.title = 'Calendar MVP — создать свой календарь';
    render(
      '<div class="page page--narrow">' +
        '<a class="link-back" href="/owners" data-nav="/owners">← Назад</a>' +
        '<h1>Создать свой календарь</h1>' +
        '<p class="subtitle">Создайте публичную страницу для записи на встречи. Для вас автоматически будут созданы два типа встреч: на 15 и 30 минут.</p>' +
        '<form id="owner-form" class="section">' +
          '<div class="field">' +
            '<label class="field__label" for="owner-name">Имя <span class="required">*</span></label>' +
            '<input class="input" id="owner-name" name="name" type="text" required maxlength="120" placeholder="Имя и фамилия" />' +
          '</div>' +
          '<div class="field">' +
            '<label class="field__label" for="owner-email">Email <span class="required">*</span></label>' +
            '<input class="input" id="owner-email" name="email" type="email" required placeholder="you@example.com" />' +
          '</div>' +
          '<div class="form-error" id="owner-form-error"></div>' +
          '<button class="btn btn--primary" type="submit">Создать календарь</button>' +
        '</form>' +
      '</div>'
    );
  }

  function submitOwnerForm(form) {
    var name = form.name.value.trim();
    var email = form.email.value.trim();
    var errorEl = document.getElementById('owner-form-error');
    errorEl.textContent = '';
    var btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.classList.add('btn--disabled');
    API.createOwner({ name: name, email: email })
      .then(function (owner) {
        Router.replace('/owners/' + owner.uuid);
      })
      .catch(function (err) {
        errorEl.textContent = errorText(err, {
          conflict: 'Email уже используется',
        });
        btn.disabled = false;
        btn.classList.remove('btn--disabled');
      });
  }

  /* ====================== Страница владельца (визард) ====================== */

  function resetWizard() {
    state.owner = null;
    state.events = [];
    state.bookings = [];
    state.event = null;
    state.slots = null;
    state.selectedDate = null;
    state.selectedSlot = null;
    state.booking = null;
    state.guestName = '';
    state.guestEmail = '';
  }

  function renderOwnerWizard(params) {
    setTheme(false);
    resetWizard();
    var uuid = params.uuid;
    document.title = 'Calendar MVP — запись на встречу';
    render(loadingHtml());
    Promise.all([API.getOwner(uuid), API.getOwnerEvents(uuid), API.getOwnerBookings(uuid)])
      .then(function (r) {
        state.owner = r[0];
        state.events = r[1].events;
        state.bookings = r[2].bookings;
        renderStep1();
      })
      .catch(function (err) {
        render(
          '<div class="page">' +
            '<a class="link-back" href="/owners" data-nav="/owners">← Назад к списку</a>' +
            '<div class="empty">' +
              (err instanceof ApiError && err.status === 404 ? 'Владелец не найден' : esc(err.message)) +
            '</div>' +
          '</div>'
        );
      });
  }

  function renderStep1() {
    document.title = state.owner.name + ' — запись на встречу';
    var tz = state.owner.settings.timezone;
    var bookingsHtml = state.bookings.length
      ? state.bookings.map(function (b) {
          return (
            '<div class="booking-row">' +
              '<span class="booking-row__main">' + esc(dateShortInTz(b.start_time, tz)) + ', ' + esc(timeInTz(b.start_time, tz)) + ' — ' + esc(b.event_name) + '</span>' +
              '<span class="booking-row__guest">' + esc(b.guest_name) + '</span>' +
            '</div>'
          );
        }).join('')
      : '<div class="empty">Пока нет бронирований</div>';

    var eventsHtml = state.events.length
      ? '<div class="card-grid card-grid--stack">' + state.events.map(eventCardHtml).join('') + '</div>'
      : '<div class="empty">Нет доступных типов встреч</div>';

    render(
      '<div class="page">' +
        '<a class="link-back" href="/owners" data-nav="/owners">← Назад к списку</a>' +
        '<div class="wizard-head">' +
          stepsHtml(1) +
          '<h1 class="wizard-head__name">' + esc(state.owner.name) + '</h1>' +
          '<p class="subtitle">Записывайтесь на встречу в часовом поясе ' + esc(tz) + '</p>' +
        '</div>' +
        '<section class="section">' +
          '<h2 class="section__title">Доступные типы встреч</h2>' +
          eventsHtml +
        '</section>' +
        '<section class="section">' +
          '<h2 class="section__title">Уже забронировано</h2>' +
          bookingsHtml +
        '</section>' +
        '<div class="row-actions">' +
          '<button class="btn btn--primary" data-action="start-booking">Забронировать время</button>' +
        '</div>' +
      '</div>'
    );
  }

  function renderStep2() {
    render(
      '<div class="page">' +
        '<div class="wizard-head">' +
          stepsHtml(2) +
          '<h1>Выберите тип встречи</h1>' +
        '</div>' +
        (state.events.length
          ? '<div class="card-grid">' + state.events.map(eventCardHtml).join('') + '</div>'
          : '<div class="empty">Нет доступных типов встреч</div>') +
        '<div class="row-actions">' +
          '<button class="btn btn--secondary" data-action="step-1">← Назад</button>' +
        '</div>' +
      '</div>'
    );
  }

  function selectEvent(eventUuid) {
    var ev = state.events.find(function (e) { return e.uuid === eventUuid; });
    if (!ev) return;
    state.event = ev;
    state.slots = null;
    state.selectedDate = null;
    state.selectedSlot = null;
    render(loadingHtml());
    API.getSlots(state.owner.uuid, ev.uuid)
      .then(function (slots) {
        state.slots = slots;
        var dates = Object.keys(slots.slots).sort();
        state.selectedDate = dates.find(function (d) { return hasAvailable(slots.slots[d]); }) || null;
        renderStep3();
      })
      .catch(function (err) {
        renderStep2Error(err);
      });
  }

  function renderStep2Error(err) {
    render(
      '<div class="page">' +
        '<div class="wizard-head">' +
          stepsHtml(2) +
          '<h1>Выберите тип встречи</h1>' +
        '</div>' +
        '<div class="empty">' + esc(err.message) + '</div>' +
        '<div class="row-actions">' +
          '<button class="btn btn--secondary" data-action="step-1">← Назад</button>' +
        '</div>' +
      '</div>'
    );
  }

  function calendarHtml() {
    var slots = state.slots;
    var start = parseDate(slots.start_date);
    var end = parseDate(slots.end_date);
    var cursor = new Date(Date.UTC(start.y, start.m - 1, start.d));
    var endDt = new Date(Date.UTC(end.y, end.m - 1, end.d));
    var days = [];
    while (cursor <= endDt) {
      days.push({ y: cursor.getUTCFullYear(), m: cursor.getUTCMonth() + 1, d: cursor.getUTCDate() });
      cursor = new Date(Date.UTC(cursor.getUTCFullYear(), cursor.getUTCMonth(), cursor.getUTCDate() + 1));
    }

    var heads = WEEKDAY_SHORT.map(function (w) {
      return '<div class="calendar__head">' + w + '</div>';
    }).join('');

    var cells = '';
    var firstDow = new Date(Date.UTC(start.y, start.m - 1, start.d)).getUTCDay();
    var lead = (firstDow + 6) % 7; // понедельник — первый день недели
    for (var i = 0; i < lead; i++) {
      cells += '<div></div>';
    }

    days.forEach(function (day) {
      var key = dateKey(day);
      var daySlots = slots.slots[key] || [];
      var available = hasAvailable(daySlots);
      var dow = new Date(Date.UTC(day.y, day.m - 1, day.d)).getUTCDay();
      var isWeekend = dow === 0 || dow === 6;
      var selected = key === state.selectedDate;
      var cls = 'calendar__day';
      if (!available) cls += ' calendar__day--disabled';
      if (isWeekend && !available) cls += ' calendar__day--weekend';
      if (selected) cls += ' calendar__day--selected';
      cells +=
        '<button class="' + cls + '" data-action="select-date" data-date="' + key + '"' +
        (available ? '' : ' disabled') + '>' + day.d + '</button>';
    });

    return (
      '<div class="calendar">' +
        '<div class="calendar__title">' + esc(formatMonthYear(slots.start_date)) + '</div>' +
        heads + cells +
      '</div>'
    );
  }

  function renderStep3() {
    var slots = state.slots;
    render(
      '<div class="page">' +
        '<div class="wizard-head">' +
          stepsHtml(3) +
          '<h1>Выберите дату и время</h1>' +
          '<p class="subtitle">' + esc(state.event.name) + ', ' + state.event.duration_minutes + ' мин · ' + esc(state.owner.settings.timezone) + '</p>' +
        '</div>' +
        '<div class="booking-layout">' +
          '<div>' + calendarHtml() + '</div>' +
          '<div>' + slotsPanelHtml() + '</div>' +
        '</div>' +
        '<div class="row-actions">' +
          '<button class="btn btn--secondary" data-action="step-2">← Назад</button>' +
        '</div>' +
      '</div>'
    );
  }

  function slotsPanelHtml() {
    if (!state.selectedDate) {
      return '<div class="empty">Нет доступных слотов в этом окне</div>';
    }
    var daySlots = state.slots.slots[state.selectedDate] || [];
    var duration = state.event.duration_minutes;
    var availableCount = daySlots.filter(function (s) { return s.status === 'available'; }).length;

    var list = daySlots.map(function (s) {
      var range = s.time + '–' + slotEnd(s.time, duration);
      if (s.status === 'available') {
        return (
          '<button class="slot slot--available" data-action="select-slot" data-time="' + esc(s.time) + '">' +
            '<span>' + range + '</span>' +
            '<span class="slot__meta">доступно</span>' +
          '</button>'
        );
      }
      var reasonLabel = s.reason === 'booked' ? 'забронировано' : 'недоступно';
      return (
        '<button class="slot slot--unavailable" disabled>' +
          '<span>' + range + '</span>' +
          '<span class="slot__meta">' + reasonLabel + '</span>' +
        '</button>'
      );
    }).join('');

    return (
      '<div class="slots-head">' +
        '<h3>Слоты ' + esc(formatDateShort(state.selectedDate)) + '</h3>' +
        '<span class="slots-count">Доступно: ' + availableCount + ' ' + plural(availableCount, 'слот', 'слота', 'слотов') + '</span>' +
      '</div>' +
      '<div class="slot-list">' +
        (list || '<div class="empty">Нет слотов в этот день</div>') +
      '</div>'
    );
  }

  function renderStep4() {
    var range = state.selectedSlot.time + '–' + slotEnd(state.selectedSlot.time, state.event.duration_minutes);
    render(
      '<div class="page">' +
        '<div class="wizard-head">' +
          stepsHtml(4) +
          '<h1>Подтвердите запись</h1>' +
        '</div>' +
        '<div class="booking-layout">' +
          '<aside class="booking-summary">' +
            '<h3 class="section__title">Резюме брони</h3>' +
            '<div class="booking-summary__list">' +
              '<div class="booking-summary__item">' +
                '<span class="booking-summary__label">Что</span>' +
                '<span class="booking-summary__value">' + esc(state.event.name) + ', ' + state.event.duration_minutes + ' мин</span>' +
              '</div>' +
              '<div class="booking-summary__item">' +
                '<span class="booking-summary__label">Когда</span>' +
                '<span class="booking-summary__value">' + esc(formatDateShort(state.selectedDate)) + ', ' + range + '</span>' +
              '</div>' +
              '<div class="booking-summary__item">' +
                '<span class="booking-summary__label">Пояс</span>' +
                '<span class="booking-summary__value">' + esc(state.owner.settings.timezone) + '</span>' +
              '</div>' +
            '</div>' +
          '</aside>' +
          '<form id="guest-form" novalidate>' +
            '<div class="field">' +
              '<label class="field__label" for="guest-name">Ваше имя <span class="required">*</span></label>' +
              '<input class="input" id="guest-name" name="guest_name" type="text" required minlength="1" maxlength="120" value="' + esc(state.guestName) + '" placeholder="Имя" />' +
            '</div>' +
            '<div class="field">' +
              '<label class="field__label" for="guest-email">Email <span class="required">*</span></label>' +
              '<input class="input" id="guest-email" name="guest_email" type="email" required value="' + esc(state.guestEmail) + '" placeholder="you@example.com" />' +
            '</div>' +
            '<div class="form-error" id="guest-form-error"></div>' +
            '<div class="row-actions">' +
              '<button class="btn btn--secondary" type="button" data-action="step-3">← Назад</button>' +
              '<button class="btn btn--primary" id="guest-submit" type="submit">Подтвердить бронирование</button>' +
            '</div>' +
          '</form>' +
        '</div>' +
      '</div>'
    );

    var form = document.getElementById('guest-form');
    var submit = document.getElementById('guest-submit');
    function update() {
      var valid = form.checkValidity();
      submit.disabled = !valid;
      submit.classList.toggle('btn--disabled', !valid);
    }
    form.addEventListener('input', update);
    update();
  }

  function submitBooking(form) {
    var name = form.guest_name.value.trim();
    var email = form.guest_email.value.trim();
    state.guestName = name;
    state.guestEmail = email;

    var submit = document.getElementById('guest-submit');
    submit.disabled = true;
    submit.classList.add('btn--disabled');
    var errorEl = document.getElementById('guest-form-error');
    errorEl.textContent = '';

    API.createBooking({
      owner_uuid: state.owner.uuid,
      event_uuid: state.event.uuid,
      guest_name: name,
      guest_email: email,
      date: state.selectedDate,
      start_time: state.selectedSlot.time,
    })
      .then(function (booking) {
        state.booking = booking;
        renderStep5();
      })
      .catch(function (err) {
        errorEl.textContent = errorText(err, {
          slot_unavailable: 'Выбранный слот уже занят. Вернитесь назад и выберите другое время.',
          overlap: 'Выбранный слот уже занят. Вернитесь назад и выберите другое время.',
          rate_limit: 'Слишком много запросов. Подождите и попробуйте снова.',
        });
        submit.disabled = false;
        submit.classList.remove('btn--disabled');
      });
  }

  function renderStep5() {
    var b = state.booking;
    var range = state.selectedSlot.time + '–' + slotEnd(state.selectedSlot.time, state.event.duration_minutes);
    render(
      '<div class="page page--narrow">' +
        '<div class="wizard-head">' +
          stepsHtml(5) +
          '<div class="success-icon">✓</div>' +
          '<h1>Встреча запланирована</h1>' +
          '<p class="subtitle">Мы отправили уведомление владельцу (mock)</p>' +
        '</div>' +
        '<div class="card">' +
          '<div class="details-list">' +
            '<div class="details-list__item">' +
              '<span class="details-list__label">Что</span>' +
              '<span class="details-list__value">' + esc(b.event_name) + ', ' + b.duration_minutes + ' мин</span>' +
            '</div>' +
            '<div class="details-list__item">' +
              '<span class="details-list__label">Когда</span>' +
              '<span class="details-list__value">' + esc(formatDateShort(state.selectedDate)) + ', ' + range + ' (' + esc(state.owner.settings.timezone) + ')</span>' +
            '</div>' +
            '<div class="details-list__item">' +
              '<span class="details-list__label">Кто</span>' +
              '<span class="details-list__value">' + esc(state.guestName) + ' — ' + esc(state.guestEmail) + '</span>' +
            '</div>' +
          '</div>' +
        '</div>' +
        '<div class="row-actions">' +
          '<button class="btn btn--primary" data-action="back-to-calendar">← Вернуться к календарю ' + esc(state.owner.name) + '</button>' +
        '</div>' +
      '</div>'
    );
  }

  function backToCalendar() {
    render(loadingHtml());
    API.getOwnerBookings(state.owner.uuid)
      .then(function (res) {
        state.bookings = res.bookings;
        state.event = null;
        state.slots = null;
        state.selectedDate = null;
        state.selectedSlot = null;
        state.booking = null;
        renderStep1();
      })
      .catch(function (err) {
        render(errorHtml(err));
      });
  }

  /* ========================= События (делегирование) ======================== */

  app.addEventListener('click', function (e) {
    var el = e.target.closest('[data-nav], [data-tab], [data-action]');
    if (!el || !app.contains(el)) return;

    if (el.hasAttribute('data-nav')) {
      e.preventDefault();
      Router.navigate(el.getAttribute('data-nav'));
      return;
    }

    if (el.hasAttribute('data-tab')) {
      state.adminTab = el.getAttribute('data-tab');
      renderAdminPage();
      return;
    }

    var action = el.getAttribute('data-action');
    switch (action) {
      case 'start-booking':
        renderStep2();
        break;
      case 'select-event':
        selectEvent(el.getAttribute('data-event-uuid'));
        break;
      case 'select-date':
        if (state.selectedDate !== el.getAttribute('data-date')) {
          state.selectedDate = el.getAttribute('data-date');
          state.selectedSlot = null;
          renderStep3();
        }
        break;
      case 'select-slot':
        state.selectedSlot = { date: state.selectedDate, time: el.getAttribute('data-time') };
        renderStep4();
        break;
      case 'step-1':
        renderStep1();
        break;
      case 'step-2':
        renderStep2();
        break;
      case 'step-3':
        renderStep3();
        break;
      case 'back-to-calendar':
        backToCalendar();
        break;
    }
  });

  app.addEventListener('submit', function (e) {
    var form = e.target;
    if (!form || form.tagName !== 'FORM') return;
    if (form.id === 'admin-event-form') {
      e.preventDefault();
      submitAdminEvent(form);
    } else if (form.id === 'owner-form') {
      e.preventDefault();
      submitOwnerForm(form);
    } else if (form.id === 'guest-form') {
      e.preventDefault();
      submitBooking(form);
    }
  });

  /* =============================== Роутинг ================================ */

  Router.register('/', renderLanding);
  Router.register('/office', renderAdmin);
  Router.register('/owners', renderOwnersList);
  Router.register('/owners/new', renderCreateOwner);
  Router.register('/owners/:uuid', renderOwnerWizard);
})();
