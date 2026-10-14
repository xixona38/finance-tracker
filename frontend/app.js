import { accountTypes, operationTypes, escapeHTML as esc, money, toMinorUnits, localDateTime } from './lib.js';
import { enhanceSelects, closeSelect } from './select.js';

const app = document.querySelector('#app');
const dialog = document.querySelector('#dialog');
const state = { user: null, accounts: [], transactions: [], view: 'overview', type: '', account: '', page: 0, next: false, busy: false, error: '', demo: false, authMode: 'login', generation: 0 };
let demoTransactions = [];
let loadSequence = 0;
let toastTimer;
let dialogTrigger;
const paths = {
  grid: 'M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h7v7h-7z',
  wallet: 'M4 5h15v4M4 5a2 2 0 0 0 0 4h17v11H4a2 2 0 0 1-2-2V7a2 2 0 0 1 2-2M17 13h4v4h-4z',
  history: 'M3 11a9 9 0 1 1 2.6 7M3 4v7h7M12 7v5l3 2', plus: 'M12 5v14M5 12h14',
  arrow: 'M7 17 17 7M7 7h10v10', income: 'M17 7 7 17M7 7v10h10', expense: 'M7 17 17 7M7 7h10v10',
  transfer: 'M4 8h16m-4-4 4 4-4 4M20 16H4m4-4-4 4 4 4',
  shield: 'M12 3 4 6v6c0 5 8 9 8 9s8-4 8-9V6zM8 12l3 3 5-6',
  logout: 'M9 4H4v16h5M10 12h11m-4-4 4 4-4 4', close: 'M6 6l12 12M6 18 18 6',
  chevron: 'm9 5 7 7-7 7', back: 'm15 5-7 7 7 7', trash: 'M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7M14 10v7',
  card: 'M3 5h18v14H3zM3 10h18M6 15h4', cash: 'M3 5h18v14H3zM9 12a3 3 0 1 0 6 0 3 3 0 1 0-6 0M6 12h.1M18 12h.1',
  savings: 'M12 3v18M7 7h7a3 3 0 0 1 0 6h-4a3 3 0 0 0 0 6h7', refresh: 'M20 8a8 8 0 0 0-14-3L3 8M3 3v5h5M4 16a8 8 0 0 0 14 3l3-3M21 21v-5h-5',
};
const icon = name => `<svg viewBox="0 0 24 24" aria-hidden="true"><path d="${paths[name] || paths.wallet}"/></svg>`;
const brand = (extra = '') => `<a href="#" class="brand ${extra}" aria-label="На главную"><img src="/favicon.svg" width="44" height="44" alt=""></a>`;
const accountName = id => state.accounts.find(a => a.id === id)?.name || 'Счёт недоступен';
const today = () => new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }).format(new Date());
const dateLabel = iso => {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? 'Дата недоступна' : new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' }).format(date);
};
function toast(message) {
  const el = document.querySelector('#toast');
  clearTimeout(toastTimer); el.textContent = message; el.hidden = false;
  toastTimer = setTimeout(() => { el.hidden = true; }, 4500);
}
function resetSession() {
  state.generation++; loadSequence++;
  Object.assign(state, { user: null, accounts: [], transactions: [], type: '', account: '', page: 0, error: '', demo: false, busy: false, view: 'overview' });
  closeDialog();
}
async function api(path, options = {}) {
  let response;
  try {
    response = await fetch(`/api/v1${path}`, {
      credentials: 'same-origin', ...options,
      headers: { ...(options.body ? { 'Content-Type': 'application/json' } : {}), ...options.headers },
    });
  } catch { throw new Error('Не удалось связаться с сервером. Проверьте подключение и запуск бэкенда.'); }
  let data = null;
  if (response.status !== 204) {
    const text = await response.text();
    try { data = JSON.parse(text); } catch { data = {}; }
  }
  if (!response.ok) {
    if (response.status === 401 && !['/login', '/register'].includes(path)) {
      const wasLoggedIn = !!state.user;
      resetSession(); renderAuth();
      if (wasLoggedIn) toast('Сессия завершилась. Войдите снова.');
    }
    const messages = { 400: 'Проверьте введённые данные.', 401: 'Неверный email или пароль.', 403: 'Запрос отклонён. Откройте приложение по его локальному адресу.', 404: 'Запись не найдена. Возможно, она уже удалена.', 409: 'Пользователь с этим email уже существует.', 413: 'Слишком большой запрос.', 429: 'Слишком много запросов. Попробуйте позже.', 500: 'Ошибка сервера. Попробуйте позже.', 502: 'Бэкенд недоступен. Запустите его командой make run.' };
    const error = new Error(messages[response.status] || 'Не удалось выполнить запрос.');
    error.status = response.status;
    throw error;
  }
  return data;
}
function renderAuth(message = '') {
  closeSelect();
  const register = state.authMode === 'register';
  app.innerHTML = `<main class="auth-page">
    <section class="auth-story">${brand()}<div class="story-copy"><div class="eyebrow">МЕНЬШЕ ШУМА. БОЛЬШЕ ЯСНОСТИ.</div><h1>Ваши деньги.<br>Ничего<br><span>лишнего.</span></h1><p>Счета, повседневные расходы и большие планы. Всё важное о ваших финансах — в одном месте.</p></div><div class="story-art" aria-hidden="true">₽</div><footer><span>01 / ЯСНОСТЬ</span><span>С вниманием к главному.</span></footer></section>
    <section class="auth-panel"><div class="auth-box"><div class="eyebrow">ЛИЧНОЕ ПРОСТРАНСТВО</div><h2>${register ? 'Начнём с чистого листа' : 'С возвращением'}</h2><p>${register ? 'Создайте аккаунт и добавьте свой первый счёт.' : 'Войдите, чтобы увидеть полную картину.'}</p>
    <div class="auth-switch"><button type="button" data-auth="login" class="${!register ? 'active' : ''}">Вход</button><button type="button" data-auth="register" class="${register ? 'active' : ''}">Регистрация</button></div>
    <form id="auth-form" class="form"><label class="field">Email<input name="email" type="email" autocomplete="username" placeholder="you@example.com" required></label><label class="field">Пароль<input name="password" type="password" autocomplete="${register ? 'new-password' : 'current-password'}" required placeholder="Не менее 15 символов"><small>От 15 до 128 символов. Можно использовать длинную фразу.</small></label>${register ? '<label class="field">Повторите пароль<input name="confirm" type="password" autocomplete="new-password" required placeholder="Ещё раз, чтобы не ошибиться"></label>' : ''}<p class="form-error" role="alert">${esc(message)}</p><button class="button primary" type="submit">${register ? 'Создать аккаунт' : 'Войти'} ${icon('arrow')}</button></form>
    <p class="auth-note">${icon('shield')} Вход через защищённую HttpOnly-сессию.</p><div class="demo-link"><button class="text-button" data-action="demo">Посмотреть интерфейс в деморежиме ${icon('chevron')}</button></div></div></section></main>`;
  document.querySelector('#auth-form').addEventListener('submit', async event => {
    event.preventDefault();
    const form = event.currentTarget;
    const values = Object.fromEntries(new FormData(form));
    const errorBox = form.querySelector('.form-error');
    const button = form.querySelector('[type=submit]');
    errorBox.textContent = '';
    if ([...values.password].length < 15 || [...values.password].length > 128) { errorBox.textContent = 'Пароль должен содержать от 15 до 128 символов.'; return; }
    if (register && values.password !== values.confirm) { errorBox.textContent = 'Пароли не совпадают.'; return; }
    button.disabled = true;
    const generation = state.generation;
    try {
      const user = await api(register ? '/register' : '/login', { method: 'POST', body: JSON.stringify({ email: values.email.trim(), password: values.password }) });
      if (generation !== state.generation || !form.isConnected) return;
      if (register) {
        state.authMode = 'login'; renderAuth();
        document.querySelector('[name=email]').value = values.email.trim();
        document.querySelector('[name=password]').focus();
        toast('Аккаунт создан. Теперь войдите.');
      } else { state.user = user; await refresh(); }
    } catch (error) { if (form.isConnected) errorBox.textContent = error.message; }
    finally { button.disabled = false; }
  });
}
function render() {
  closeSelect();
  if (!state.user) return renderAuth();
  const titles = { overview: ['Обзор', ''], accounts: ['Мои счета', 'Каждому счёту — своё место.'], history: ['История операций', 'Всё, что приходит и уходит.'] };
  const [title, subtitle] = titles[state.view];
  app.innerHTML = `<div class="shell"><aside class="sidebar">${brand()}<div class="nav-label">ВАШЕ ПРОСТРАНСТВО</div><nav class="nav" aria-label="Основная навигация">${[['overview', 'grid', 'Обзор'], ['accounts', 'wallet', 'Счета'], ['history', 'history', 'Операции']].map(([view, symbol, label]) => `<button data-view="${view}" class="${state.view === view ? 'active' : ''}" ${state.view === view ? 'aria-current="page"' : ''}>${icon(symbol)}${label}${state.view === view ? '<span class="nav-dot"></span>' : ''}</button>`).join('')}</nav><div class="sidebar-bottom"><div class="profile"><span class="avatar">${esc(state.user.email[0].toUpperCase())}</span><div class="profile-info"><strong>${esc(state.user.email)}</strong><span>${state.demo ? 'Деморежим' : 'Личный аккаунт'}</span></div><button class="icon-button" data-action="logout" aria-label="Выйти" title="Выйти">${icon('logout')}</button></div></div></aside>
    <main class="workspace"><header class="topbar"><div class="topbar-left">${icon('grid')} Пространство <span>/</span> <span>${state.view === 'overview' ? 'Обзор' : title}</span></div>${brand('mobile-brand')}<div class="topbar-right"><span>${today()}</span><button class="icon-button" data-action="refresh" aria-label="Обновить данные" title="Обновить" ${state.busy ? 'disabled' : ''}>${icon('refresh')}</button><button class="icon-button mobile-logout" data-action="logout" aria-label="Выйти" title="Выйти">${icon('logout')}</button></div></header>
    ${state.demo ? '<div class="demo-banner"><span>Деморежим · пример данных, изменения не сохраняются на сервере</span><button data-action="logout">К входу →</button></div>' : ''}
    <div class="page-heading ${state.view === 'overview' ? 'overview-actions' : ''}">${state.view === 'overview' ? '' : `<div><h1>${title}</h1><p>${subtitle}</p></div>`}<div class="heading-actions"><button class="button ghost" data-action="transfer" ${state.busy ? 'disabled' : ''}>${icon('transfer')} Перевод</button><button class="button primary" data-action="operation" ${state.busy ? 'disabled' : ''}>${icon('plus')} Добавить операцию</button></div></div>
    ${state.error ? `<div class="notice" role="alert">${esc(state.error)} <button class="text-button" data-action="refresh">Повторить</button></div>` : ''}
    ${state.view === 'overview' ? overviewHTML() : ''}
    ${state.view !== 'history' ? accountsHTML() : ''}
    ${state.view !== 'accounts' ? historyHTML() : ''}
    <footer class="page-footer"><span>Личные финансы / Без лишнего</span><span>${state.busy ? 'Обновляем данные…' : state.error ? 'Не удалось обновить' : 'Все суммы в рублях'}</span></footer></main></div>`;
  enhanceSelects(app);
}
function overviewHTML() {
  const total = state.accounts.reduce((sum, a) => sum + a.balance, 0);
  const positive = state.accounts.reduce((sum, a) => sum + Math.max(0, a.balance), 0);
  const groups = Object.entries(accountTypes).map(([type, label]) => ({ label, amount: state.accounts.filter(a => a.type === type).reduce((sum, a) => sum + a.balance, 0) })).filter(g => g.amount !== 0);
  return `<section class="overview" aria-label="Общий баланс"><div class="balance-card"><div class="balance-label">${icon('wallet')} Общий баланс</div><div class="balance-number">${state.busy && !state.accounts.length ? '—' : money(total)}</div><div class="balance-footer"><span>На всех ваших счетах</span><span>RUB · ₽</span></div></div><div class="distribution"><div class="distribution-title">Где ваши деньги <span>По типам счетов</span></div>${groups.length ? groups.map(g => `<div class="distribution-row"><span>${g.label}</span><strong>${money(g.amount)}</strong></div><div class="bar"><meter min="0" max="${positive || 1}" value="${Math.max(0, g.amount)}" aria-label="${g.label}"></meter></div>`).join('') : `<div class="empty">${icon('wallet')}Здесь появится распределение<br>баланса по вашим счетам.</div>`}</div></section>`;
}
function accountsHTML() {
  return `<section class="${state.view === 'accounts' ? 'all-accounts' : ''}"><div class="section-head"><h2>Мои счета <span class="count">${state.accounts.length}</span></h2><button class="text-button" data-action="account">${icon('plus')} Новый счёт</button></div><div class="accounts">${state.accounts.map(a => `<button class="account-card" data-account="${a.id}" title="Показать операции счёта"><div class="account-top"><span class="icon-box">${icon(a.type === 'credit' ? 'card' : a.type)}</span><span class="account-tag">RUB ${icon('chevron')}</span></div><h3>${esc(a.name)}</h3><p>${accountTypes[a.type] || 'Счёт'}</p><div class="account-amount">${money(a.balance)}</div></button>`).join('')}${!state.accounts.length || state.view === 'accounts' ? `<button class="account-card account-add" data-action="account">${icon('plus')}<span>${state.accounts.length ? 'Добавить счёт' : 'Добавить первый счёт'}</span><small>Карта, наличные или накопления</small></button>` : ''}</div></section>`;
}
function historyHTML() {
  const options = state.accounts.map(a => `<option value="${a.id}" ${state.account === String(a.id) ? 'selected' : ''}>${esc(a.name)}</option>`).join('');
  return `<section><div class="section-head"><h2>${state.view === 'overview' ? 'Последние операции' : 'Операции'}</h2><span class="eyebrow">История движения денег</span></div><div class="history" aria-busy="${state.busy}"><div class="history-toolbar"><div class="tabs" role="group" aria-label="Тип операции">${[['', 'Все'], ['expense', 'Расходы'], ['income', 'Доходы'], ['transfer', 'Переводы']].map(([value, label]) => `<button class="tab ${state.type === value ? 'active' : ''}" data-type="${value}" aria-pressed="${state.type === value}">${label}</button>`).join('')}</div><select class="filter-select" id="account-filter" aria-label="Фильтр по счёту"><option value="">Все счета</option>${options}</select></div>
    ${state.busy ? '<div class="empty" role="status">Загружаем операции…</div>' : state.transactions.length ? `<div class="table-scroll"><table><thead><tr><th>Операция</th><th class="account-cell">Счёт</th><th class="date-cell">Дата</th><th class="amount">Сумма</th><th><span class="muted"> </span></th></tr></thead><tbody>${state.transactions.map(t => {
      const name = t.description || operationTypes[t.type];
      const accounts = t.type === 'transfer' ? `${accountName(t.from_acc_id)} → ${accountName(t.to_acc_id)}` : accountName(t.from_acc_id ?? t.to_acc_id);
      return `<tr><td><div class="operation"><span class="icon-box ${t.type}">${icon(t.type)}</span><div><strong title="${esc(name)}">${esc(name)}</strong><small>${operationTypes[t.type]}</small></div></div></td><td class="account-cell">${esc(accounts)}</td><td class="date-cell">${dateLabel(t.occurred_at)}</td><td class="amount ${t.type}">${t.type === 'income' ? '+' : t.type === 'expense' ? '−' : ''}${money(t.amount)}</td><td class="delete-cell"><button class="icon-button" data-delete="${t.id}" aria-label="Удалить операцию ${esc(name)}" title="Удалить">${icon('trash')}</button></td></tr>`;
    }).join('')}</tbody></table></div>` : `<div class="empty">${icon('history')}<strong>${state.type || state.account ? 'Ничего не найдено' : 'История начинается с вас'}</strong>${state.type || state.account ? 'Попробуйте другой счёт или тип операции.' : 'Добавьте первый доход, расход или перевод.'}${!state.type && !state.account ? '<br><button class="button ghost" data-action="operation">Добавить операцию</button>' : ''}</div>`}
    <div class="pagination"><span>${state.transactions.length ? `${state.page * 20 + 1}–${state.page * 20 + state.transactions.length} · по 20 на странице` : 'Нет операций'}</span><div class="pagination-controls"><button class="icon-button" data-action="previous" aria-label="Предыдущая страница" ${!state.page || state.busy ? 'disabled' : ''}>${icon('back')}</button><span>${state.page + 1}</span><button class="icon-button" data-action="next" aria-label="Следующая страница" ${!state.next || state.busy ? 'disabled' : ''}>${icon('chevron')}</button></div></div></div></section>`;
}
async function refresh() {
  const sequence = ++loadSequence;
  const generation = state.generation;
  state.busy = true; state.error = ''; render();
  try {
    let accounts, list;
    if (state.demo) {
      accounts = state.accounts;
      list = demoTransactions.filter(t => (!state.type || t.type === state.type) && (!state.account || t.from_acc_id === Number(state.account) || t.to_acc_id === Number(state.account))).sort((a, b) => b.occurred_at.localeCompare(a.occurred_at) || b.id - a.id).slice(state.page * 20, state.page * 20 + 21);
    } else {
      const query = new URLSearchParams({ limit: '21', offset: String(state.page * 20) });
      if (state.type) query.set('type', state.type);
      if (state.account) query.set('account_id', state.account);
      [accounts, list] = await Promise.all([api('/accounts'), api(`/transactions?${query}`)]);
    }
    if (sequence !== loadSequence || generation !== state.generation) return;
    state.accounts = accounts || []; state.next = list.length > 20; state.transactions = list.slice(0, 20);
    if (!list.length && state.page > 0) { state.page--; return refresh(); }
  } catch (error) {
    if (sequence !== loadSequence || generation !== state.generation) return;
    state.error = error.message; state.transactions = []; state.next = false;
  } finally {
    if (sequence === loadSequence && generation === state.generation) { state.busy = false; render(); }
  }
}
function showDialog(title, content) {
  closeSelect();
  dialogTrigger = document.activeElement;
  dialog.innerHTML = `<div class="dialog-head"><h2 id="dialog-title">${title}</h2><button class="icon-button" data-close aria-label="Закрыть">${icon('close')}</button></div>${content}`;
  dialog.showModal();
  enhanceSelects(dialog);
}
function closeDialog() {
  closeSelect();
  if (dialog.open) dialog.close();
  dialogTrigger?.isConnected && dialogTrigger.focus();
}
function bindForm(callback) {
  const form = dialog.querySelector('form');
  form.addEventListener('submit', async event => {
    event.preventDefault();
    if (form.dataset.saving) return;
    const errorBox = form.querySelector('.form-error');
    const values = Object.fromEntries(new FormData(form));
    errorBox.textContent = '';
    form.dataset.saving = 'true';
    const controls = [...dialog.querySelectorAll('button, input, select')];
    controls.forEach(c => { c.disabled = true; });
    try { await callback(values); }
    catch (error) { if (form.isConnected && dialog.open) errorBox.textContent = error.message; }
    finally { delete form.dataset.saving; controls.forEach(c => { c.disabled = false; }); }
  });
}
function openAccount() {
  showDialog('Новый счёт', `<form class="form"><label class="field">Название<input name="name" required maxlength="100" placeholder="Например, основная карта" autofocus></label><div class="form-row"><label class="field">Тип счёта<select name="type">${Object.entries(accountTypes).map(([value, label]) => `<option value="${value}">${label}</option>`).join('')}</select></label><label class="field">Валюта<select name="currency"><option value="RUB">Рубли · RUB</option></select></label></div><label class="field">Начальный баланс, ₽<input name="balance" inputmode="decimal" value="0" required><small>Сколько денег на счёте прямо сейчас. Можно указать отрицательный остаток.</small></label><p class="form-error" role="alert"></p><button class="button primary" type="submit">Создать счёт</button></form>`);
  bindForm(async values => {
    if (!values.name.trim()) throw new Error('Введите название счёта.');
    const payload = { name: values.name.trim(), type: values.type, currency: 'RUB', initial_balance: toMinorUnits(values.balance, true) };
    if (state.demo) state.accounts.push({ ...payload, balance: payload.initial_balance, id: Date.now() });
    else await api('/accounts', { method: 'POST', body: JSON.stringify(payload) });
    closeDialog(); toast('Счёт создан'); await refresh();
  });
}
function openOperation(type = 'expense') {
  if (!state.accounts.length) { toast('Сначала добавьте счёт'); return openAccount(); }
  let currentType = type;
  const options = state.accounts.map(a => `<option value="${a.id}">${esc(a.name)} · ${money(a.balance)}</option>`).join('');
  showDialog('Новая операция', `<form class="form"><div class="segmented" role="group" aria-label="Вид операции">${Object.entries(operationTypes).map(([value, label]) => `<button type="button" data-operation="${value}">${label}</button>`).join('')}</div><label class="field">Сумма, ₽<input name="amount" inputmode="decimal" placeholder="0,00" required autofocus></label><label class="field" id="source-field">Списать со счёта<select name="from">${options}</select></label><label class="field" id="destination-field">Зачислить на счёт<select name="to">${options}</select></label><label class="field">Дата и время<input name="date" type="datetime-local" value="${localDateTime()}" required></label><label class="field">Описание <span class="muted">Необязательно</span><input name="description" maxlength="500" placeholder="На что потратили или откуда поступили деньги"></label><p class="form-summary" id="operation-hint"></p><p class="form-error" role="alert"></p><button class="button primary" type="submit">Сохранить операцию</button></form>`);
  if (state.accounts.length > 1) {
    const destination = dialog.querySelector('[name=to]');
    destination.value = String(state.accounts[1].id);
    destination.dispatchEvent(new Event('change', { bubbles: true }));
  }
  function selectType(value) {
    closeSelect();
    currentType = value;
    dialog.querySelectorAll('[data-operation]').forEach(b => { b.classList.toggle('active', b.dataset.operation === value); b.setAttribute('aria-pressed', String(b.dataset.operation === value)); });
    dialog.querySelector('#source-field').hidden = value === 'income';
    dialog.querySelector('#destination-field').hidden = value === 'expense';
    dialog.querySelector('#operation-hint').textContent = value === 'transfer' ? 'Перевод меняет балансы двух счетов, но не общую сумму денег.' : value === 'income' ? 'Поступление увеличит баланс выбранного счёта.' : 'Расход уменьшит баланс выбранного счёта.';
  }
  dialog.querySelectorAll('[data-operation]').forEach(b => b.addEventListener('click', () => selectType(b.dataset.operation)));
  selectType(type);
  bindForm(async values => {
    const from = currentType === 'income' ? null : Number(values.from);
    const to = currentType === 'expense' ? null : Number(values.to);
    if (currentType === 'transfer' && from === to) throw new Error('Для перевода нужны два разных счёта.');
    const date = new Date(values.date);
    if (Number.isNaN(date.getTime())) throw new Error('Укажите корректную дату.');
    const payload = { type: currentType, from_acc_id: from, to_acc_id: to, amount: toMinorUnits(values.amount), description: values.description.trim(), occurred_at: date.toISOString() };
    if (state.demo) {
      const transaction = { ...payload, id: Date.now() };
      demoTransactions.push(transaction); applyDemoBalance(transaction, 1);
    } else await api('/transactions', { method: 'POST', body: JSON.stringify(payload) });
    state.page = 0; closeDialog(); toast('Операция добавлена'); await refresh();
  });
}
function openDelete(id) {
  const transaction = state.transactions.find(t => t.id === id);
  if (!transaction) return;
  showDialog('Удалить операцию?', `<p class="dialog-description">${esc(transaction.description || operationTypes[transaction.type])}<br><strong>${money(transaction.amount)}</strong><br>Балансы счетов будут пересчитаны. Отменить удаление нельзя.</p><form><p class="form-error" role="alert"></p><div class="dialog-actions"><button type="button" class="button ghost" data-close>Оставить</button><button type="submit" class="button danger">Удалить операцию</button></div></form>`);
  bindForm(async () => {
    if (state.demo) { demoTransactions = demoTransactions.filter(t => t.id !== id); applyDemoBalance(transaction, -1); }
    else await api(`/transactions/${id}`, { method: 'DELETE' });
    closeDialog(); toast('Операция удалена'); await refresh();
  });
}
function applyDemoBalance(t, sign) {
  for (const a of state.accounts) {
    if (a.id === t.from_acc_id) a.balance -= sign * t.amount;
    if (a.id === t.to_acc_id) a.balance += sign * t.amount;
  }
}
function startDemo() {
  resetSession();
  state.demo = true; state.user = { email: 'demo@example.com' };
  state.accounts = [ { id: 1, name: 'Основная карта', type: 'card', balance: 8435000 }, { id: 2, name: 'На большие планы', type: 'savings', balance: 18000000 }, { id: 3, name: 'Наличные', type: 'cash', balance: 750000 } ];
  const examples = [ ['expense', 1, null, 248000, 'Продукты на неделю'], ['income', null, 1, 9500000, 'Зарплата'], ['transfer', 1, 2, 2000000, 'Пополнение накоплений'], ['expense', 1, null, 39000, 'Кофе и прогулка'], ['expense', 3, null, 180000, 'Книги для вдохновения'], ['income', null, 1, 1250000, 'Дополнительный проект'] ];
  demoTransactions = examples.map(([type, from_acc_id, to_acc_id, amount, description], i) => ({ id: i + 1, type, from_acc_id, to_acc_id, amount, description, occurred_at: new Date(Date.now() - i * 86400000).toISOString() }));
  refresh();
}
app.addEventListener('change', event => {
  if (event.target.id === 'account-filter') {
    state.account = event.target.value; state.page = 0;
    refresh().then(() => app.querySelector('#account-filter + .custom-select .select-trigger')?.focus({ preventScroll: true }));
  }
});
app.addEventListener('click', async event => {
  const b = event.target.closest('button, a.brand');
  if (!b || b.disabled) return;
  if (b.matches('a.brand')) { event.preventDefault(); if (state.user) { state.view = 'overview'; render(); } return; }
  if (b.dataset.auth) { state.generation++; state.authMode = b.dataset.auth; return renderAuth(); }
  if (b.dataset.view) { state.view = b.dataset.view; return render(); }
  if ('type' in b.dataset) { state.type = b.dataset.type; state.page = 0; return refresh(); }
  if (b.dataset.account) { state.account = b.dataset.account; state.view = 'history'; state.page = 0; return refresh(); }
  if (b.dataset.delete) return openDelete(Number(b.dataset.delete));
  switch (b.dataset.action) {
    case 'demo': startDemo(); break;
    case 'account': openAccount(); break;
    case 'operation': openOperation(); break;
    case 'transfer': openOperation('transfer'); break;
    case 'refresh': await refresh(); break;
    case 'previous': state.page--; await refresh(); break;
    case 'next': state.page++; await refresh(); break;
    case 'logout':
      b.disabled = true;
      try {
        if (!state.demo) await api('/logout', { method: 'POST' });
        resetSession(); renderAuth();
      } catch (error) { toast(error.message); }
      finally { b.disabled = false; }
      break;
  }
});
dialog.addEventListener('click', event => { if (event.target.closest('[data-close]') && !dialog.querySelector('[data-saving]')) closeDialog(); });
dialog.addEventListener('cancel', event => { if (dialog.querySelector('[data-saving]')) event.preventDefault(); else closeSelect(); });
async function boot() {
  try { state.user = await api('/me'); await refresh(); }
  catch (error) { renderAuth(error.status === 401 ? '' : error.message); }
}
boot();
