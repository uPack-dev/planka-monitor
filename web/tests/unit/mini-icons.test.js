import { existsSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const projectRoot = process.cwd();

const iconPath = (name) =>
  resolve(projectRoot, 'app', 'assets', 'icons', `${name}.svg`);

const source = (path) => readFileSync(resolve(projectRoot, path), 'utf8');

describe('mini app Figma icon mapping', () => {
  it('ships dedicated Figma action icon assets', () => {
    const icons = [
      'workspace-mini',
      'clock-mini',
      'message-filled-mini',
      'check-filled-mini',
      'bell-filled-mini',
      'back-mini',
      'mention-mini',
    ];

    for (const icon of icons) {
      expect(existsSync(iconPath(icon))).toBe(true);
      expect(source(iconPath(icon))).toContain('currentColor');
    }
  });

  it('uses Figma action icons in the task action sheet', () => {
    const taskActionSheet = source('app/components/mini/TaskActionSheet.vue');

    expect(taskActionSheet).toContain('<CIcon name="bell-filled-mini" />');
    expect(taskActionSheet).toContain('<CIcon name="check-filled-mini" />');
    expect(taskActionSheet).toContain('<CIcon name="workspace-mini" />');
    expect(taskActionSheet).toContain('<CIcon name="message-filled-mini" />');
    expect(taskActionSheet).toContain('<CIcon name="clock-mini" />');
    expect(taskActionSheet).not.toContain('<CIcon name="link" />');
    expect(taskActionSheet).not.toContain('<CIcon name="calendar-mini" />');
  });

  it('uses Figma action icons in the card detail screen', () => {
    const cardSheet = source('app/components/mini/CardSheet.vue');

    expect(cardSheet).toContain('<CIcon name="back-mini" />');
    expect(cardSheet).toContain('<CIcon name="check-filled-mini" />');
    expect(cardSheet).toContain('<CIcon name="message-filled-mini" />');
    expect(cardSheet).toContain('<CIcon name="workspace-mini" />');
    expect(cardSheet).toContain('<CIcon name="mention-mini" />');
    expect(cardSheet).toContain(
      "'card-detail__action--active': isComposerOpen",
    );
    expect(cardSheet).toContain('@click="toggleCommentComposer"');
    expect(cardSheet).not.toContain('<CIcon name="chevron-left" />');
    expect(cardSheet).not.toContain('<CIcon name="link" />');
    expect(cardSheet).not.toContain('card-detail__action--blue');
  });

  it('keeps calendar task markers above day backgrounds', () => {
    const tasksPage = source('app/components/mini/screens/TasksPage.vue');
    const markerStyles = tasksPage.match(
      /&__day-marker \{(?<styles>[\s\S]*?)\n {2}\}/,
    )?.groups?.styles;

    expect(tasksPage).toContain('class="calendar-panel__day-marker"');
    expect(markerStyles).toContain('z-index: 2;');
  });
});
