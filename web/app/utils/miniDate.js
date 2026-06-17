const UK_LOCALE = 'uk-UA';
const MS_PER_HOUR = 36e5;
const MS_PER_MINUTE = 60000;

export function parseValidDate(value) {
  if (!value) return null;
  const date = value instanceof Date ? new Date(value) : new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

export function dateKey(value) {
  const date = parseValidDate(value);
  if (!date) return '';

  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

export function parseDateKey(key) {
  return new Date(`${key}T00:00:00`);
}

export function startOfDay(value = new Date()) {
  const date = parseValidDate(value) || new Date();
  date.setHours(0, 0, 0, 0);
  return date;
}

export function endOfDay(value = new Date()) {
  const date = parseValidDate(value) || new Date();
  date.setHours(23, 59, 59, 999);
  return date;
}

export function startOfMonth(value) {
  const date = parseValidDate(value) || new Date();
  return new Date(date.getFullYear(), date.getMonth(), 1);
}

export function endOfMonth(value) {
  const date = parseValidDate(value) || new Date();
  return new Date(date.getFullYear(), date.getMonth() + 1, 0, 23, 59, 59, 999);
}

export function addDays(value, days) {
  const date = parseValidDate(value) || new Date();
  date.setDate(date.getDate() + days);
  return date;
}

export function addMinutes(value, minutes) {
  const date = parseValidDate(value) || new Date();
  date.setMinutes(date.getMinutes() + minutes);
  return date;
}

export function setTime(value, time) {
  const [hours, minutes] = String(time).split(':').map(Number);
  const date = parseValidDate(value) || new Date();
  date.setHours(hours || 0, minutes || 0, 0, 0);
  return date;
}

export function mondayIndex(value) {
  const date = parseValidDate(value) || new Date();
  return (date.getDay() + 6) % 7;
}

export function getCalendarYearRangeStart(value) {
  const date = parseValidDate(value) || new Date();
  return Math.floor(date.getFullYear() / 10) * 10;
}

export function formatClockTime(value) {
  const date = parseValidDate(value);
  if (!date) return '';
  return date.toLocaleTimeString(UK_LOCALE, {
    hour: '2-digit',
    minute: '2-digit',
  });
}

export function formatCalendarDate(value) {
  const date = parseValidDate(value);
  if (!date) return '';
  return date.toLocaleDateString(UK_LOCALE, {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  });
}

export function dayLabel(
  value,
  { now = new Date(), capitalized = false, includeYesterday = true } = {},
) {
  const date = parseValidDate(value);
  if (!date) return '';

  const labels = capitalized
    ? { today: 'Сьогодні', tomorrow: 'Завтра', yesterday: 'Вчора' }
    : { today: 'сьогодні', tomorrow: 'завтра', yesterday: 'вчора' };
  const key = dateKey(date);
  const today = startOfDay(now);

  if (key === dateKey(today)) return labels.today;
  if (key === dateKey(addDays(today, 1))) return labels.tomorrow;
  if (includeYesterday && key === dateKey(addDays(today, -1))) {
    return labels.yesterday;
  }
  return formatCalendarDate(date);
}

export function relativeDeadlineLabel(value, now = new Date()) {
  const date = parseValidDate(value);
  if (!date) return '';

  const diff = date.getTime() - now.getTime();
  const absHours = Math.max(1, Math.round(Math.abs(diff) / MS_PER_HOUR));

  if (diff < 0) {
    return absHours >= 24
      ? `прострочено на ${Math.round(absHours / 24)} дн.`
      : `прострочено на ${absHours} год.`;
  }

  return absHours >= 24
    ? `через ${Math.round(absHours / 24)} дн.`
    : `через ${absHours} год.`;
}

export function relativePastLabel(value, now = new Date()) {
  const date = parseValidDate(value);
  if (!date) return '';

  const diff = date.getTime() - now.getTime();
  const absMinutes = Math.max(1, Math.round(Math.abs(diff) / MS_PER_MINUTE));

  if (absMinutes < 60) return `${absMinutes} хв. тому`;
  const absHours = Math.round(absMinutes / 60);
  if (absHours < 24) return `${absHours} год. тому`;
  return `${Math.round(absHours / 24)} дн. тому`;
}
