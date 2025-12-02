<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { PhDatabase, PhBroom, PhHardDrive, PhCalendarX, PhEyeSlash } from '@phosphor-icons/vue';
import type { SettingsData } from '@/types/settings';

const { t } = useI18n();

interface Props {
  settings: SettingsData;
}

defineProps<Props>();
</script>

<template>
  <div class="setting-group">
    <label
      class="font-semibold mb-2 sm:mb-3 text-text-secondary uppercase text-xs tracking-wider flex items-center gap-2"
    >
      <PhDatabase :size="14" class="sm:w-4 sm:h-4" />
      {{ t('database') }}
    </label>
    <div class="setting-item">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhBroom :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">{{ t('autoCleanup') }}</div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('autoCleanupDesc') }}
          </div>
        </div>
      </div>
      <input type="checkbox" v-model="settings.auto_cleanup_enabled" class="toggle" />
    </div>

    <div
      v-if="settings.auto_cleanup_enabled"
      class="ml-2 sm:ml-4 mt-2 sm:mt-3 space-y-2 sm:space-y-3 border-l-2 border-border pl-2 sm:pl-4"
    >
      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhHardDrive :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('maxCacheSize') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('maxCacheSizeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            type="number"
            v-model="settings.max_cache_size_mb"
            min="1"
            max="1000"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
          />
          <span class="text-xs sm:text-sm text-text-secondary">MB</span>
        </div>
      </div>

      <div class="sub-setting-item">
        <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
          <PhCalendarX :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
          <div class="flex-1 min-w-0">
            <div class="font-medium mb-0 sm:mb-1 text-sm">{{ t('maxArticleAge') }}</div>
            <div class="text-xs text-text-secondary hidden sm:block">
              {{ t('maxArticleAgeDesc') }}
            </div>
          </div>
        </div>
        <div class="flex items-center gap-1 sm:gap-2 shrink-0">
          <input
            type="number"
            v-model="settings.max_article_age_days"
            min="1"
            max="365"
            class="input-field w-14 sm:w-20 text-center text-xs sm:text-sm"
          />
          <span class="text-xs sm:text-sm text-text-secondary">{{ t('days') }}</span>
        </div>
      </div>
    </div>

    <div class="setting-item mt-2 sm:mt-3">
      <div class="flex-1 flex items-center sm:items-start gap-2 sm:gap-3 min-w-0">
        <PhEyeSlash :size="20" class="text-text-secondary mt-0.5 shrink-0 sm:w-6 sm:h-6" />
        <div class="flex-1 min-w-0">
          <div class="font-medium mb-0 sm:mb-1 text-sm sm:text-base">
            {{ t('showHiddenArticles') }}
          </div>
          <div class="text-xs text-text-secondary hidden sm:block">
            {{ t('showHiddenArticlesDesc') }}
          </div>
        </div>
      </div>
      <input type="checkbox" v-model="settings.show_hidden_articles" class="toggle" />
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
.sub-setting-item {
  @apply flex items-center sm:items-start justify-between gap-2 sm:gap-4 p-2 sm:p-2.5 rounded-md bg-bg-tertiary;
}
</style>
