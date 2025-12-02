<script setup lang="ts">
import { PhSpinnerGap } from '@phosphor-icons/vue';

interface Props {
  articleContent: string;
  isTranslatingContent: boolean;
}

defineProps<Props>();
</script>

<template>
  <!-- Content display with inline translations -->
  <div v-if="articleContent">
    <div
      class="prose prose-sm sm:prose-lg max-w-none text-text-primary prose-content"
      v-html="articleContent"
    ></div>
    <!-- Translation loading indicator -->
    <div v-if="isTranslatingContent" class="flex items-center gap-2 mt-4 text-text-secondary">
      <PhSpinnerGap :size="16" class="animate-spin" />
      <span class="text-sm">Translating content...</span>
    </div>
  </div>

  <!-- No content available -->
  <div v-else class="text-center text-text-secondary py-6 sm:py-8">
    <PhArticle :size="48" class="mb-2 sm:mb-3 opacity-50 mx-auto sm:w-16 sm:h-16" />
    <p class="text-sm sm:text-base">No content available</p>
  </div>
</template>
