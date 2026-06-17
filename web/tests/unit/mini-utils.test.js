import { describe, expect, it } from 'vitest';
import {
  addDays,
  dateKey,
  dayLabel,
  formatClockTime,
  parseValidDate,
  relativeDeadlineLabel,
  relativePastLabel,
  startOfDay,
} from '../../app/utils/miniDate.js';
import {
  displayUserName,
  normalizeUsername,
  userInitials,
} from '../../app/utils/miniUsers.js';
import { escapeHtml, formatRichText } from '../../app/utils/richText.js';

describe('mini app utility modules', () => {
  it('normalizes dates and labels calendar days', () => {
    const now = new Date('2026-06-15T10:30:00');
    const today = new Date('2026-06-15T18:45:00');
    const tomorrow = addDays(startOfDay(now), 1);

    expect(parseValidDate('not-a-date')).toBeNull();
    expect(dateKey(today)).toBe('2026-06-15');
    expect(formatClockTime(today)).toBe('18:45');
    expect(dayLabel(today, { now })).toBe('сьогодні');
    expect(dayLabel(tomorrow, { now, capitalized: true })).toBe('Завтра');
  });

  it('formats relative mini app time labels', () => {
    const now = new Date('2026-06-15T10:00:00');

    expect(relativeDeadlineLabel('2026-06-15T12:00:00', now)).toBe(
      'через 2 год.',
    );
    expect(relativeDeadlineLabel('2026-06-14T10:00:00', now)).toBe(
      'прострочено на 1 дн.',
    );
    expect(relativePastLabel('2026-06-15T09:30:00', now)).toBe('30 хв. тому');
  });

  it('formats mini app users', () => {
    expect(normalizeUsername(' @Yan ')).toBe('yan');
    expect(displayUserName({ username: 'yan', name: 'Ян' })).toBe('Ян');
    expect(userInitials({ name: 'Yan Dev' })).toBe('YD');
    expect(userInitials('')).toBe('?');
  });

  it('formats a safe rich text subset', () => {
    expect(escapeHtml('<script>"x"</script>')).toBe(
      '&lt;script&gt;&quot;x&quot;&lt;/script&gt;',
    );
    expect(
      formatRichText('## Title\n- **item**\n- `code`\n\n@[Yan](u1)', {
        mentionClass: 'mention',
      }),
    ).toBe(
      '<h4>Title</h4><ul><li><strong>item</strong></li><li><code>code</code></li></ul><p><span class="mention">@Yan</span></p>',
    );
  });
});
