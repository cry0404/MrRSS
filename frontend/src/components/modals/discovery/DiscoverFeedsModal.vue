<script setup lang="ts">
import { watch, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhX } from '@phosphor-icons/vue';
import type { Feed } from '@/types/models';
import DiscoveredFeedItem from './DiscoveredFeedItem.vue';
import DiscoveryProgress from './DiscoveryProgress.vue';
import { useModalClose } from '@/composables/ui/useModalClose';
import { useFeedDiscovery } from '@/composables/discovery/useFeedDiscovery';
import { useFeedSubscription } from '@/composables/discovery/useFeedSubscription';

const { t } = useI18n();

// Modal close handling
useModalClose(() => close());

interface Props {
  feed: Feed;
  show: boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  close: [];
}>();

// Use discovery composable
const {
  isDiscovering,
  discoveredFeeds,
  errorMessage,
  progressMessage,
  progressDetail,
  progressCounts,
  startDiscovery,
  cleanup: cleanupDiscovery,
} = useFeedDiscovery(props.feed);

// Use subscription composable
const {
  selectedFeeds,
  isSubscribing,
  hasSelection,
  allSelected,
  toggleFeedSelection,
  selectAll,
  subscribeSelected,
} = useFeedSubscription(props.feed, discoveredFeeds);

function close() {
  // Clear polling interval if active
  cleanupDiscovery();
  emit('close');
}

// Auto-start discovery when component is mounted
onMounted(() => {
  if (props.show) {
    startDiscovery();
  }
});

// Watch for modal opening and trigger discovery (for when modal is reused)
watch(
  () => props.show,
  (newShow, oldShow) => {
    if (newShow && !oldShow) {
      startDiscovery();
    }
  }
);

// Cleanup on unmount
onUnmounted(() => {
  cleanupDiscovery();
});
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
    @click.self="close"
    data-modal-open="true"
  >
    <div
      class="bg-bg-primary w-full max-w-4xl max-h-[90vh] rounded-2xl shadow-2xl border border-border flex flex-col"
    >
      <!-- Header -->
      <div
        class="flex justify-between items-center p-6 border-b border-border bg-gradient-to-r from-accent/5 to-transparent"
      >
        <div>
          <h2 class="text-xl font-bold text-text-primary">{{ t('discoverFeeds') }}</h2>
          <p class="text-sm text-text-secondary mt-1">{{ t('fromFeed') }}: {{ feed.title }}</p>
        </div>
        <button @click="close" class="p-2 hover:bg-bg-tertiary rounded-lg transition-colors">
          <PhX :size="24" class="text-text-secondary" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6">
        <!-- Loading State -->
        <DiscoveryProgress
          v-if="isDiscovering"
          :progress-message="progressMessage"
          :progress-detail="progressDetail"
          :progress-counts="progressCounts"
        />

        <!-- Error State -->
        <div
          v-else-if="errorMessage"
          class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 text-red-600 dark:text-red-400"
        >
          {{ errorMessage }}
        </div>

        <!-- Results -->
        <div v-else-if="discoveredFeeds.length > 0">
          <div class="mb-4 flex items-center justify-between bg-bg-secondary rounded-lg p-3">
            <p class="text-sm font-medium text-text-primary">
              {{ t('foundFeeds', { count: discoveredFeeds.length }) }}
            </p>
            <button
              @click="selectAll"
              class="text-sm text-accent hover:text-accent-hover font-medium px-3 py-1 rounded hover:bg-accent/10 transition-colors"
            >
              {{ allSelected ? t('deselectAll') : t('selectAll') }}
            </button>
          </div>

          <div class="space-y-3">
            <DiscoveredFeedItem
              v-for="(feed, index) in discoveredFeeds"
              :key="index"
              :feed="feed"
              :is-selected="selectedFeeds.has(index)"
              @toggle="toggleFeedSelection(index)"
            />
          </div>
        </div>

        <!-- Initial State (should not be visible as discovery auto-starts) -->
        <div v-else class="text-center py-16">
          <div
            class="w-16 h-16 border-4 border-accent border-t-transparent rounded-full animate-spin mx-auto mb-4"
          ></div>
          <p class="text-text-secondary text-lg">{{ t('preparing') }}...</p>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-between items-center p-6 border-t border-border bg-bg-secondary/50">
        <button @click="close" class="btn-secondary" :disabled="isSubscribing">
          {{ t('cancel') }}
        </button>
        <button
          @click="subscribeSelected"
          :disabled="!hasSelection || isSubscribing"
          :class="[
            'btn-primary flex items-center gap-2',
            (!hasSelection || isSubscribing) && 'opacity-50 cursor-not-allowed',
          ]"
        >
          <div
            v-if="isSubscribing"
            class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"
          ></div>
          {{ isSubscribing ? t('subscribing') : t('subscribeSelected') }}
          <span
            v-if="hasSelection && !isSubscribing"
            class="bg-white/20 px-2 py-0.5 rounded-full text-sm"
            >({{ selectedFeeds.size }})</span
          >
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.btn-primary {
  @apply px-6 py-2.5 bg-accent text-white rounded-lg hover:bg-accent-hover transition-all font-medium shadow-sm hover:shadow-md;
}

.btn-secondary {
  @apply px-6 py-2.5 bg-bg-tertiary text-text-primary rounded-lg hover:opacity-80 transition-all font-medium;
}
</style>
