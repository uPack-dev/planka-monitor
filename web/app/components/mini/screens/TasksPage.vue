<template>
  <section class="screen tasks-screen">
    <section v-if="taskView !== 'calendar'" class="tasks-hero">
      <header>
        <span>{{ currentWeekday }}</span>
        <time :datetime="selectedDate">{{ currentDateLabel }}</time>
      </header>
      <h1>{{ greetingLabel }}</h1>
      <div class="tasks-hero__stats">
        <span class="tasks-hero__pill tasks-hero__pill--danger">
          <CIcon name="alert-circle-mini" />
          {{ taskStats.overdue }}
        </span>
        <span class="tasks-hero__pill tasks-hero__pill--today">
          На сьогодні: {{ taskStats.today }}
        </span>
        <span class="tasks-hero__pill tasks-hero__pill--done">
          <CIcon name="calendar-mini" />
          {{ taskStats.planned }}
        </span>
      </div>
    </section>

    <div v-if="taskView === 'calendar'" class="calendar-layout">
      <section
        class="calendar-panel calendar-layout__month"
        aria-label="Календар задач"
      >
        <header class="calendar-panel__top">
          <button
            class="calendar-panel__arrow"
            type="button"
            aria-label="Попередній місяць"
            @click="$emit('change-calendar-month', -1)"
          >
            <CIcon name="chevron-left" />
          </button>
          <div
            ref="calendarPeriodControls"
            class="calendar-panel__period"
            aria-label="Період календаря"
          >
            <button
              class="calendar-panel__period-control"
              :class="{
                'calendar-panel__period-control--active':
                  calendarPickerMode === 'month',
              }"
              type="button"
              aria-haspopup="listbox"
              :aria-expanded="calendarPickerMode === 'month'"
              aria-label="Обрати місяць"
              @click="$emit('toggle-calendar-picker', 'month')"
            >
              <strong>{{ calendarMonthLabel }}</strong>
            </button>
            <button
              class="calendar-panel__period-control"
              :class="{
                'calendar-panel__period-control--active':
                  calendarPickerMode === 'year',
              }"
              type="button"
              aria-haspopup="listbox"
              :aria-expanded="calendarPickerMode === 'year'"
              aria-label="Обрати рік"
              @click="$emit('toggle-calendar-picker', 'year')"
            >
              <span>{{ calendarYearLabel }}</span>
            </button>
          </div>
          <div
            v-if="calendarPickerMode"
            ref="calendarPicker"
            class="calendar-panel__picker"
          >
            <div
              v-if="calendarPickerMode === 'month'"
              class="calendar-panel__picker-grid calendar-panel__picker-grid--months"
              role="listbox"
              aria-label="Місяць"
            >
              <button
                v-for="month in calendarMonthOptions"
                :key="month.value"
                class="calendar-panel__picker-option"
                :class="{
                  'calendar-panel__picker-option--active':
                    calendarViewDateObject.getMonth() === month.value,
                }"
                type="button"
                role="option"
                :aria-selected="
                  calendarViewDateObject.getMonth() === month.value
                "
                @click="$emit('select-calendar-month', month.value)"
              >
                {{ month.label }}
              </button>
            </div>
            <div v-else class="calendar-panel__year-picker">
              <div class="calendar-panel__picker-nav">
                <button
                  class="calendar-panel__picker-arrow"
                  type="button"
                  aria-label="Попередні роки"
                  @click="$emit('change-calendar-year-page', -1)"
                >
                  <CIcon name="chevron-left" />
                </button>
                <strong>{{ calendarYearRangeLabel }}</strong>
                <button
                  class="calendar-panel__picker-arrow calendar-panel__picker-arrow--next"
                  type="button"
                  aria-label="Наступні роки"
                  @click="$emit('change-calendar-year-page', 1)"
                >
                  <CIcon name="chevron-left" />
                </button>
              </div>
              <div
                class="calendar-panel__picker-grid calendar-panel__picker-grid--years"
                role="listbox"
                aria-label="Рік"
              >
                <button
                  v-for="year in calendarYearOptions"
                  :key="year"
                  class="calendar-panel__picker-option"
                  :class="{
                    'calendar-panel__picker-option--active':
                      calendarViewDateObject.getFullYear() === year,
                  }"
                  type="button"
                  role="option"
                  :aria-selected="calendarViewDateObject.getFullYear() === year"
                  @click="$emit('select-calendar-year', year)"
                >
                  {{ year }}
                </button>
              </div>
            </div>
          </div>
          <button
            class="calendar-panel__arrow calendar-panel__arrow--next"
            type="button"
            aria-label="Наступний місяць"
            @click="$emit('change-calendar-month', 1)"
          >
            <CIcon name="chevron-left" />
          </button>
        </header>
        <div class="calendar-panel__weekdays" aria-hidden="true">
          <span v-for="weekday in calendarDayNames" :key="weekday">
            {{ weekday }}
          </span>
        </div>
        <div class="calendar-panel__grid">
          <button
            v-for="day in monthCalendarDays"
            :key="day.key"
            class="calendar-panel__day"
            type="button"
            :class="{
              'calendar-panel__day--active': selectedDate === day.key,
              'calendar-panel__day--outside': !day.isCurrentMonth,
              'calendar-panel__day--today': day.isToday,
              'calendar-panel__day--has-tasks': day.taskCount,
            }"
            @click="$emit('select-date', day.key)"
          >
            <span class="calendar-panel__day-number">
              {{ day.day }}
            </span>
            <i v-if="day.taskCount" class="calendar-panel__day-marker" />
          </button>
        </div>
      </section>

      <section class="calendar-layout__agenda" aria-live="polite">
        <p class="calendar-day-label">
          {{ selectedDateLabel }}
        </p>

        <div v-if="isLoading && !visibleTaskItems.length" class="loader">
          Оновлюємо задачі...
        </div>

        <MiniEmptyState
          v-else-if="!visibleTaskItems.length"
          icon="check"
          title="Усе зроблено."
          text="Прострочених немає, на сьогодні нічого не призначено. Можна видихнути або зайнятись творчим."
        />

        <div v-else class="calendar-task-list">
          <MiniTaskCard
            v-for="(item, itemIndex) in visibleTaskItems"
            :key="item.id"
            :item="item"
            :style="{ '--motion-delay': `${itemIndex * 20}ms` }"
            @open="$emit('open-card', $event)"
            @complete="$emit('complete', $event)"
            @menu="$emit('menu', $event)"
          />
        </div>
      </section>
    </div>

    <template v-else>
      <div v-if="isLoading && !visibleTaskItems.length" class="loader">
        Оновлюємо задачі...
      </div>

      <MiniEmptyState
        v-else-if="!visibleTaskItems.length"
        icon="check"
        title="Усе зроблено."
        text="Прострочених немає, на сьогодні нічого не призначено. Можна видихнути або зайнятись творчим."
      />

      <div v-else class="timeline">
        <section
          v-for="(group, groupIndex) in taskGroups"
          :key="group.key"
          class="timeline__group"
          :class="`timeline__group--${group.tone}`"
          :style="{ '--motion-delay': `${groupIndex * 24}ms` }"
        >
          <header>
            <span>{{ group.label }}</span>
          </header>
          <MiniTaskCard
            v-for="(item, itemIndex) in group.items"
            :key="item.id"
            :item="item"
            :style="{ '--motion-delay': `${itemIndex * 20}ms` }"
            @open="$emit('open-card', $event)"
            @complete="$emit('complete', $event)"
            @menu="$emit('menu', $event)"
          />
        </section>
      </div>
    </template>
  </section>
</template>

<script setup>
const props = defineProps({
  taskView: {
    type: String,
    required: true,
  },
  currentWeekday: {
    type: String,
    required: true,
  },
  selectedDate: {
    type: String,
    required: true,
  },
  currentDateLabel: {
    type: String,
    required: true,
  },
  greetingLabel: {
    type: String,
    required: true,
  },
  taskStats: {
    type: Object,
    required: true,
  },
  calendarPickerMode: {
    type: String,
    required: true,
  },
  calendarMonthLabel: {
    type: String,
    required: true,
  },
  calendarYearLabel: {
    type: String,
    required: true,
  },
  calendarMonthOptions: {
    type: Array,
    required: true,
  },
  calendarViewDateObject: {
    type: Date,
    required: true,
  },
  calendarYearRangeLabel: {
    type: String,
    required: true,
  },
  calendarYearOptions: {
    type: Array,
    required: true,
  },
  calendarDayNames: {
    type: Array,
    required: true,
  },
  monthCalendarDays: {
    type: Array,
    required: true,
  },
  selectedDateLabel: {
    type: String,
    required: true,
  },
  isLoading: {
    type: Boolean,
    required: true,
  },
  visibleTaskItems: {
    type: Array,
    required: true,
  },
  taskGroups: {
    type: Array,
    required: true,
  },
});

const emit = defineEmits([
  'change-calendar-month',
  'toggle-calendar-picker',
  'close-calendar-picker',
  'select-calendar-month',
  'change-calendar-year-page',
  'select-calendar-year',
  'select-date',
  'open-card',
  'complete',
  'menu',
]);

const calendarPeriodControls = ref(null);
const calendarPicker = ref(null);

onMounted(() => {
  document.addEventListener('pointerdown', handleDocumentPointerDown, true);
  document.addEventListener('keydown', handleDocumentKeydown, true);
});

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', handleDocumentPointerDown, true);
  document.removeEventListener('keydown', handleDocumentKeydown, true);
});

function handleDocumentPointerDown(event) {
  if (!props.calendarPickerMode) return;
  if (
    containsTarget(calendarPeriodControls.value, event.target) ||
    containsTarget(calendarPicker.value, event.target)
  ) {
    return;
  }

  emit('close-calendar-picker');
}

function handleDocumentKeydown(event) {
  if (!props.calendarPickerMode || event.key !== 'Escape') return;

  emit('close-calendar-picker');
}

function containsTarget(element, target) {
  return Boolean(element && target && element.contains(target));
}
</script>

<style scoped lang="scss">
.screen {
  display: grid;
  gap: 14px;
  align-content: start;
  min-height: max-content;
}

.loader {
  padding: 26px;
  font-size: 14px;
  font-weight: 700;
  color: $annonce-color-ink-muted;
  text-align: center;
}

.tasks-screen .loader {
  display: grid;
  place-items: center;
  min-height: 470px;
  padding: 168px 24px 70px;
}

.tasks-hero {
  position: relative;
  height: 129px;
  padding: 15px;
  color: $annonce-color-text;
  background-color: #265577;
  border-radius: 14px;
  box-shadow: 0 14px 30px rgba(#0b304b, 0.12);
  animation: mini-rise-in 240ms $ease-out both;

  &__stats {
    position: absolute;
    top: 81px;
    right: 15px;
    left: 15px;
    display: grid;
    grid-template-columns: 80px 1fr 80px;
    gap: 10px;
    align-items: center;
    height: 33px;
  }

  &__pill {
    display: inline-flex;
    gap: 5px;
    align-items: center;
    justify-content: center;
    min-width: 0;
    height: 33px;
    padding: 0 15px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 14.5px;
    font-weight: 600;
    line-height: 21px;
    white-space: nowrap;
    border: 1.5px solid transparent;
    border-radius: 360px;
    transition:
      transform $time-fast $ease-out,
      box-shadow $time-fast $ease-out;
    animation: mini-rise-in 220ms $ease-out both;

    :deep(svg) {
      flex: 0 0 20px;
      width: 20px;
      height: 20px;
    }

    &--danger {
      color: #a3263e;
      background-color: #ffe7ed;
      border-color: #f1b6c4;
      animation-delay: 35ms;
    }

    &--today {
      color: $annonce-color-yellow-action;
      background-color: #fff5db;
      border-color: $annonce-color-yellow-action;
      animation-delay: 20ms;
    }

    &--done {
      color: $annonce-color-navy;
      background-color: #e7f5ff;
      border-color: #5995c1;
      animation-delay: 15ms;
    }
  }

  header {
    position: absolute;
    top: 15px;
    right: 15px;
    left: 15px;
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: space-between;
    font-size: 14.5px;
    font-weight: 600;
    line-height: 20px;
    color: rgba($annonce-color-text, 0.7);
  }

  header span {
    text-transform: capitalize;
  }

  h1 {
    position: absolute;
    top: 41px;
    right: 15px;
    left: 15px;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-family: $annonce-font-brand;
    font-size: 24px;
    font-weight: 400;
    line-height: 30px;
    text-align: center;
    white-space: nowrap;
  }
}

.calendar-layout {
  display: grid;
  gap: 14px;
  width: 100%;
  animation: mini-rise-in 220ms $ease-out both;

  @include media-breakpoint-up(md) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 24px;
    align-items: start;
    max-width: 1040px;
    margin: 0 auto;

    &__agenda {
      padding-top: 6px;
    }

    &__agenda .loader {
      min-height: 180px;
      padding: 64px 24px 32px;
    }

    &__agenda :deep(.empty-state) {
      min-height: 0;
      padding: 36px 24px;
    }
  }

  &__agenda {
    display: grid;
    gap: 12px;
    align-content: start;
    min-width: 0;
  }
}

.calendar-panel {
  box-sizing: border-box;
  display: grid;
  width: 100%;
  min-height: 380px;
  padding: 18px;
  background-color: #e9eff3;
  border: 1px solid #d4e2ea;
  border-radius: 14px;
  box-shadow: 0 10px 28px rgba(#0b304b, 0.08);

  @include media-breakpoint-up(md) {
    min-height: 0;
    padding: 20px;
  }

  &__top {
    position: relative;
    display: grid;
    grid-template-columns: 40px minmax(0, 1fr) 40px;
    align-items: center;
    height: 40px;
    margin-bottom: 9px;
  }

  &__arrow {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    color: #f4fafd;
    background-color: #265577;
    border-radius: 50%;
    box-shadow: 0 1px 0.5px rgba(#000e33, 0.05);
    transition:
      background-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;

    :deep(svg) {
      width: 16px;
      height: 16px;
    }

    &--next {
      rotate: 180deg;
    }
  }

  &__period {
    display: inline-flex;
    gap: 10px;
    align-items: center;
    justify-self: center;
    min-width: 0;
    max-width: 100%;
  }

  &__period-control {
    display: grid;
    place-items: center;
    min-height: 40px;
    padding: 0 12px;
    color: #182a33;
    background-color: $color-white;
    border: 1px solid transparent;
    border-radius: 6px;
    box-shadow: 0 1px 0.5px rgba(#000e33, 0.05);
    transition:
      border-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;
  }

  &__period-control strong,
  &__period-control span {
    font-size: 18px;
    font-weight: 600;
    line-height: 1;
  }

  &__period-control--active {
    border-color: #ffc261;
    box-shadow:
      0 1px 0.5px rgba(#000e33, 0.05),
      0 0 0 2px rgba(#ffc261, 0.3);
  }

  &__picker {
    position: absolute;
    top: calc(100% + 8px);
    left: 50%;
    z-index: 5;
    width: min(100%, 320px);
    padding: 10px;
    background-color: $color-white;
    border: 1px solid #d4e2ea;
    border-radius: 12px;
    box-shadow: 0 14px 32px rgba(#000e33, 0.18);
    translate: -50% 0;
    animation: mini-rise-in 180ms $ease-out both;
  }

  &__picker-grid {
    display: grid;
    gap: 6px;

    &--months {
      grid-template-columns: repeat(3, minmax(0, 1fr));
    }

    &--years {
      grid-template-columns: repeat(5, minmax(0, 1fr));
    }
  }

  &__picker-option {
    min-height: 38px;
    padding: 0 8px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 13px;
    font-weight: 700;
    line-height: 1;
    color: #182a33;
    white-space: nowrap;
    background-color: #eef4f7;
    border: 1px solid transparent;
    border-radius: 8px;
    transition:
      background-color $time-fast $ease-out,
      border-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    &--active {
      background-color: #ffc261;
      border-color: #ffc261;
    }
  }

  &__year-picker {
    display: grid;
    gap: 8px;
  }

  &__picker-nav {
    display: grid;
    grid-template-columns: 34px minmax(0, 1fr) 34px;
    gap: 8px;
    align-items: center;

    strong {
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 13px;
      font-weight: 800;
      line-height: 1;
      color: #182a33;
      text-align: center;
      white-space: nowrap;
    }
  }

  &__picker-arrow {
    display: grid;
    place-items: center;
    width: 34px;
    height: 34px;
    color: #f4fafd;
    background-color: #265577;
    border-radius: 50%;
    transition:
      background-color $time-fast $ease-out,
      transform $time-fast $ease-out;

    :deep(svg) {
      width: 14px;
      height: 14px;
    }

    &--next {
      rotate: 180deg;
    }
  }

  &__weekdays,
  &__grid {
    display: grid;
    grid-template-columns: repeat(7, minmax(0, 1fr));
    column-gap: 6px;
  }

  &__weekdays {
    font-size: 14px;
    font-weight: 700;
    color: #182a33;
  }

  &__weekdays span {
    display: grid;
    place-items: center;
    height: 44px;
  }

  &__grid {
    row-gap: 6px;
  }

  &__day {
    position: relative;
    display: grid;
    place-items: center;
    width: 100%;
    height: auto;
    min-height: 47px;
    aspect-ratio: 1;
    padding: 0;
    color: #182a33;
    background-color: transparent;
    transition: transform $time-fast $ease-out;
  }

  &__day-number {
    position: relative;
    z-index: 1;
    display: grid;
    place-items: center;
    width: 100%;
    height: 100%;
    font-size: 14px;
    font-weight: 600;
    line-height: 1;
    background-color: $color-white;
    border: 1px solid transparent;
    border-radius: 6px;
    box-shadow: 0 1px 0.5px rgba(#000e33, 0.05);
    transition:
      background-color $time-fast $ease-out,
      border-color $time-fast $ease-out,
      box-shadow $time-fast $ease-out,
      transform $time-fast $ease-out;

    @include media-breakpoint-up(md) {
      font-size: 15px;
    }
  }

  &__day--outside {
    color: rgba(#001754, 0.25);

    .calendar-panel__day-number {
      background-color: transparent;
      border-color: transparent;
      box-shadow: none;
    }
  }

  &__day--today:not(.calendar-panel__day--active) {
    .calendar-panel__day-number {
      border-color: #265577;
    }
  }

  &__day--active {
    color: #f4fafd;

    .calendar-panel__day-number {
      background-color: #ffc261;
      border-color: #ffc261;
    }
  }

  &__day-marker {
    position: absolute;
    top: 3px;
    left: 50%;
    z-index: 2;
    width: 28px;
    height: 2px;
    content: '';
    background-color: #265577;
    border-radius: 2px;
    translate: -50% 0;
  }

  &__day--active &__day-marker {
    background-color: #f4fafd;
  }
}

.calendar-day-label {
  font-size: 13px;
  font-weight: 700;
  line-height: 1.25;
  color: #082d3d;
  text-align: center;

  @include media-breakpoint-up(md) {
    text-align: left;
  }
}

.calendar-task-list {
  display: grid;
  gap: 8px;
}

.timeline {
  position: relative;
  display: grid;
  gap: 14px;
  padding-left: 18px;

  &::before {
    position: absolute;
    top: 14px;
    bottom: 14px;
    left: 3px;
    width: 4px;
    content: '';
    background-color: #d5e4ec;
    border-radius: 3px;
  }

  &__group {
    --timeline-color: #184566;
    --timeline-text-color: #f4fafd;

    position: relative;
    display: grid;
    gap: 8px;
    animation: mini-rise-in 210ms $ease-out both;
    animation-delay: var(--motion-delay, 0ms);

    &::before {
      position: absolute;
      top: 13px;
      left: -18px;
      width: 8px;
      height: 8px;
      content: '';
      background-color: var(--timeline-color);
      border: 2px solid $annonce-color-cream;
      border-radius: 50%;
    }

    &::after {
      position: absolute;
      top: 16px;
      right: -14px;
      left: -10px;
      height: 2px;
      content: '';
      background-color: var(--timeline-color);
    }

    &--overdue {
      --timeline-color: #a3263e;
      --timeline-text-color: #ffffff;
    }

    &--today {
      --timeline-color: #ffc261;
      --timeline-text-color: #082d3d;
    }

    &--future {
      --timeline-color: #184566;
      --timeline-text-color: #ffffff;
    }

    &--undated {
      --timeline-color: #5d7180;
      --timeline-text-color: #ffffff;
    }
  }

  header {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: max-content;
    max-width: 100%;
    min-height: 28px;
    padding: 0 12px;
    overflow: hidden;
    font-size: 12px;
    color: var(--timeline-text-color);
    background-color: var(--timeline-color);
    border-radius: 14px;

    span {
      overflow: hidden;
      text-overflow: ellipsis;
      font-weight: 700;
      text-transform: capitalize;
      white-space: nowrap;
    }
  }

  .task-card {
    position: relative;
    z-index: 1;
  }
}

@media (hover: hover) {
  .tasks-hero__pill:hover,
  .calendar-panel__period-control:hover,
  .calendar-panel__picker-option:hover,
  .calendar-panel__day:hover {
    transform: translate3d(0, -1px, 0);
  }

  .calendar-panel__arrow:hover,
  .calendar-panel__picker-arrow:hover {
    background-color: $annonce-color-blue;
    box-shadow: 0 8px 18px rgba(#0b304b, 0.16);
    transform: translate3d(0, -1px, 0);
  }

  .calendar-panel__day:hover .calendar-panel__day-number {
    box-shadow: 0 8px 18px rgba(#0b304b, 0.1);
  }
}

@media (prefers-reduced-motion: reduce) {
  .tasks-hero,
  .tasks-hero__pill,
  .calendar-layout,
  .calendar-panel__picker,
  .calendar-panel__day-marker,
  .timeline__group {
    animation: none;
  }

  .tasks-hero__pill,
  .calendar-panel__arrow,
  .calendar-panel__period-control,
  .calendar-panel__picker-option,
  .calendar-panel__picker-arrow,
  .calendar-panel__day,
  .calendar-panel__day-number {
    transition-duration: 1ms;
  }
}
</style>
