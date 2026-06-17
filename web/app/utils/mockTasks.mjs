import {
  addDays,
  addMinutes,
  endOfDay,
  parseValidDate,
  setTime,
  startOfDay,
} from './miniDate.js';

export const MIN_DESIGN_TASK_COUNT = 8;

const MOCK_TASKS_OFF_VALUES = new Set(['0', 'false', 'off', 'none']);
const MOCK_TASKS_ON_VALUES = new Set(['1', 'true', 'on', 'auto']);

const BOARD_FIXTURES = [
  { name: 'OptTrans Cargo', color: '#ffc261' },
  { name: 'Vovk Brand', color: '#d94d67' },
  { name: 'Scrynya.Market', color: '#2e8c5a' },
  { name: 'uPack.studio', color: '#184566' },
  { name: 'Annonce Flow', color: '#52a8c8' },
];

const MEMBER_FIXTURES = [
  {
    id: 'mock-member-maria',
    username: 'maria',
    initials: 'M',
    color: '#ffc261',
  },
  { id: 'mock-member-yan', username: 'yan', initials: 'Y', color: '#265577' },
  {
    id: 'mock-member-julia',
    username: 'julia',
    initials: 'J',
    color: '#5995c1',
  },
];

const TITLE_FIXTURES = [
  'Презентація сторінки послуг',
  'Відповісти клієнту по правкам',
  'Lorem ipsum dolor sir amet Lorem ipsum dolor sir amet',
  'Підготувати матеріали для запуску',
  'Перевірити задачі перед релізом',
  'Оновити дедлайни в дорожній карті',
  'Зібрати фідбек по новому екрану',
  'Узгодити календар публікацій',
  'Перенести задачі в актуальний спринт',
  'Закрити правки після ревʼю',
];

const DATE_PATTERNS = [
  { dayOffset: -2, time: '14:30' },
  { dayOffset: 0, time: '16:00' },
  { dayOffset: 0, time: '19:00' },
  { dayOffset: 1, time: '09:00' },
  { dayOffset: 2, time: '15:00' },
  { dayOffset: 2, time: '18:00' },
  { dayOffset: 2, time: '20:00' },
  { dayOffset: 6, time: '11:30' },
  { dayOffset: 8, time: '10:00' },
  { dayOffset: 10, time: '17:30' },
];

export function ensureMockTaskCoverage(items = [], options = {}) {
  const minCount = normalizeCount(options.minCount, MIN_DESIGN_TASK_COUNT);

  if (items.length >= minCount) {
    return sortTasks(items);
  }

  const existingIds = new Set(items.map((item) => item.id));
  const mocks = generateMockTasks({
    ...options,
    count: minCount,
  }).filter((item) => !existingIds.has(item.id));

  return sortTasks([...items, ...mocks].slice(0, minCount));
}

export function isMockTasksEnabled(value) {
  const mode = String(value ?? '')
    .trim()
    .toLowerCase();

  if (MOCK_TASKS_OFF_VALUES.has(mode)) return false;
  return MOCK_TASKS_ON_VALUES.has(mode);
}

export function generateMockTasks(options = {}) {
  const count = normalizeCount(options.count, MIN_DESIGN_TASK_COUNT);
  const anchor = startOfDay(
    parseValidDate(options.selectedDate || options.now),
  );
  const rangeStart = parseValidDate(options.from) || addDays(anchor, -2);
  const rangeEnd = parseValidDate(options.to) || addDays(anchor, 14);
  const normalizedStart = startOfDay(rangeStart);
  let normalizedEnd = endOfDay(rangeEnd);
  if (normalizedEnd < normalizedStart) {
    normalizedEnd = endOfDay(addDays(normalizedStart, 14));
  }
  const dates = buildCandidateDates(
    anchor,
    normalizedStart,
    normalizedEnd,
    count,
  );

  return dates
    .slice(0, count)
    .map((dueAt, index) => createMockTask(dueAt, index));
}

export function mockTaskToCard(item) {
  const members = item.members || [];
  const firstTaskId = item.taskId || `${item.cardId}-task-1`;
  const now = new Date();

  return {
    id: item.cardId,
    title: item.title,
    boardName: item.boardName,
    members,
    isDueCompleted: false,
    isMock: true,
    dueAt: item.dueAt,
    cardWorkspaceUrl: item.cardWorkspaceUrl || item.workspaceUrl || '#',
    description:
      'Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.',
    tasks: [
      {
        id: firstTaskId,
        name: item.kind === 'task' ? item.title : 'Зібрати референси',
        isCompleted: false,
        isMock: true,
        assignee: members[0],
      },
      {
        id: `${firstTaskId}-draft`,
        name: 'Зробити перший варіант',
        isCompleted: true,
        isMock: true,
        assignee: members[1],
      },
      {
        id: `${firstTaskId}-final`,
        name: 'Зробити фінальну презентацію та показати клієнту',
        isCompleted: false,
        isMock: true,
        assignee: members[2] || members[0],
      },
      {
        id: `${firstTaskId}-fixes`,
        name: 'Зробити правки',
        isCompleted: false,
        isMock: true,
        assignee: members[0],
      },
    ],
    comments: [
      {
        id: `${item.cardId}-comment-own-2`,
        text: 'Зрозуміла. Зроблю до 14:00)',
        createdAt: addMinutes(now, -3).toISOString(),
        author: MEMBER_FIXTURES[1],
      },
      {
        id: `${item.cardId}-comment-other`,
        text: 'Робимо у фірмових кольорах, без яскравих градієнтів.',
        createdAt: addMinutes(now, -4).toISOString(),
        author: MEMBER_FIXTURES[2],
      },
      {
        id: `${item.cardId}-comment-own-1`,
        text: 'В яких кольорах робимо?',
        createdAt: addMinutes(now, -5).toISOString(),
        author: MEMBER_FIXTURES[1],
      },
    ],
  };
}

function createMockTask(dueAt, index) {
  const board = BOARD_FIXTURES[index % BOARD_FIXTURES.length];
  const members = [
    MEMBER_FIXTURES[index % MEMBER_FIXTURES.length],
    MEMBER_FIXTURES[(index + 1) % MEMBER_FIXTURES.length],
    MEMBER_FIXTURES[(index + 2) % MEMBER_FIXTURES.length],
  ].slice(0, index % 3 === 0 ? 2 : 3);
  const kind = index % 4 === 0 ? 'card' : 'task';
  const cardId = `mock-card-${index + 1}`;
  const taskId = `mock-task-${index + 1}`;

  return {
    id: kind === 'card' ? `mock:card:${index + 1}` : `mock:task:${index + 1}`,
    kind,
    cardId,
    taskId,
    title: TITLE_FIXTURES[index % TITLE_FIXTURES.length],
    boardName: board.name,
    boardColor: board.color,
    dueAt: dueAt.toISOString(),
    members,
    cardWorkspaceUrl: '#',
    workspaceUrl: '#',
    isMock: true,
  };
}

function buildCandidateDates(anchor, rangeStart, rangeEnd, count) {
  const dates = DATE_PATTERNS.map((pattern) =>
    setTime(addDays(anchor, pattern.dayOffset), pattern.time),
  ).filter((date) => date >= rangeStart && date <= rangeEnd);
  const fallbackStart =
    addDays(anchor, -2) >= rangeStart ? addDays(anchor, -2) : rangeStart;
  const fallbackEnd =
    addDays(anchor, 14) <= rangeEnd ? addDays(anchor, 14) : rangeEnd;

  const totalDays = Math.max(
    1,
    Math.ceil((fallbackEnd.getTime() - fallbackStart.getTime()) / 86400000),
  );

  for (let index = 0; dates.length < count; index += 1) {
    const dayOffset = index % totalDays;
    const hour = 9 + ((index * 3) % 10);
    const minute = index % 2 ? 30 : 0;
    const date = new Date(fallbackStart);
    date.setDate(fallbackStart.getDate() + dayOffset);
    date.setHours(hour, minute, 0, 0);
    if (date <= rangeEnd) {
      dates.push(date);
    }
  }

  return sortDates(dedupeDates(dates));
}

function dedupeDates(dates) {
  const seen = new Set();
  return dates.filter((date) => {
    const key = date.toISOString();
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

function sortTasks(items) {
  return [...items].sort((a, b) => {
    const aTime = parseSortableTime(a.dueAt);
    const bTime = parseSortableTime(b.dueAt);
    return aTime - bTime;
  });
}

function sortDates(dates) {
  return [...dates].sort((a, b) => a.getTime() - b.getTime());
}

function normalizeCount(value, fallback) {
  const count = Number(value);
  return Number.isFinite(count) && count > 0 ? Math.round(count) : fallback;
}

function parseSortableTime(value) {
  const date = parseValidDate(value);
  return date ? date.getTime() : Number.POSITIVE_INFINITY;
}
