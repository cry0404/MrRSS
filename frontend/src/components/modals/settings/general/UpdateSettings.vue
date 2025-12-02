<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhArrowClockwise, PhClock, PhCalendarCheck, PhPower } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';
import { formatRelativeTime } from '@/utils/date';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

const props = defineProps<Props>();

// Format last update time using shared utility
function formatLastUpdate(timestamp: string): string {
  return formatRelativeTime(timestamp, props.settings.language, t);
}
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhArrowClockwise :size="14" class="sm:w-4 sm:h-4" />
      {{ t('updates') }}
    </label>
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhClock :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('autoUpdateInterval') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('autoUpdateIntervalDesc') }}
          </div>
        </div>
      </div>
      <input
        type="number"
        v-model="settings.update_interval"
        min="1"
        class="input-field w-16 sm:w-20 text-center text-xs sm:text-sm"
      />
    </div>

    <!-- Last update time - read-only info display -->
    <div class="info-display mt-2 sm:mt-3">
      <div class="flex items-center gap-2">
        <PhCalendarCheck :size="18" class="text-text-secondary shrink-0 sm:w-5 sm:h-5" />
        <div class="flex-1 min-w-0">
          <div class="text-xs sm:text-sm text-text-secondary truncate">
            {{ t('lastArticleUpdate') }}
          </div>
        </div>
        <div class="text-xs sm:text-sm font-medium text-accent shrink-0">
          {{ formatLastUpdate(settings.last_article_update) }}
        </div>
      </div>
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhPower :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('startupOnBoot') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('startupOnBootDesc') }}
          </div>
        </div>
      </div>
      <input type="checkbox" v-model="settings.startup_on_boot" class="toggle" />
    </div>
  </div>
</template>

<style scoped>
.input-field {
  @apply p-1.5 sm:p-2.5 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors;
}
.toggle {
  @apply w-10 h-5 appearance-none bg-bg-tertiary rounded-full relative cursor-pointer border border-border transition-colors checked:bg-accent checked:border-accent shrink-0;
}
.toggle::after {
  content: '';
  @apply absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full shadow-sm transition-transform;
}
.toggle:checked::after {
  transform: translateX(20px);
}
.setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-3 rounded-lg bg-bg-secondary border border-border;
}
.info-display {
  @apply px-2 sm:px-3 py-1.5 sm:py-2 rounded-lg border border-border;
  background-color: rgba(233, 236, 239, 0.3);
}
.dark-mode .info-display {
  background-color: rgba(45, 45, 45, 0.3);
}
</style>
