export const accountTypes = { card: 'Банковская карта', cash: 'Наличные', savings: 'Накопления', credit: 'Кредитный счёт' };
export const operationTypes = { expense: 'Расход', income: 'Доход', transfer: 'Перевод' };
export const escapeHTML = value => String(value ?? '').replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[char]));
const formatter = new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', minimumFractionDigits: 2 });
export function money(value) {
  if (!Number.isSafeInteger(value)) return 'Сумма вне диапазона';
  return formatter.format(value / 100);
}
// Parse decimal rubles without multiplying a floating-point fraction.
export function toMinorUnits(raw, allowNegative = false) {
  const value = String(raw).trim().replace(',', '.');
  if (!/^-?\d+(?:\.\d{1,2})?$/.test(value)) throw new Error('Укажите сумму в рублях, не больше двух знаков после запятой.');
  const negative = value.startsWith('-');
  const [whole, fraction = ''] = value.replace('-', '').split('.');
  const result = (BigInt(whole) * 100n + BigInt(fraction.padEnd(2, '0'))) * (negative ? -1n : 1n);
  if (result > BigInt(Number.MAX_SAFE_INTEGER) || result < BigInt(Number.MIN_SAFE_INTEGER)) throw new Error('Слишком большая сумма.');
  if (!allowNegative && result <= 0n) throw new Error('Сумма должна быть больше нуля.');
  return Number(result);
}
export function localDateTime(date = new Date()) {
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000).toISOString().slice(0, 16);
}
