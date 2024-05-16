<script lang="ts" setup>
import { Icon } from '@iconify/vue';
import { useRouter } from 'vue-router';

const props = defineProps({
  isRefreshing: Boolean
})

const router = useRouter();

const goToSettings = () => {
  router.push('/settings');
}
</script>

<template>
  <nav>
    <div class="util">
      <button class="btn" :title="$t('Home.RssListButton.filter')" disabled>
        <Icon icon="material-symbols:filter-list" />
      </button>
    </div>
    <div class="function">
      <button class="btn" @click="$emit('handle-click-refresh')" :title="$t('Home.RssListButton.refresh')">
        <Icon v-if="!isRefreshing" icon="material-symbols:refresh" />
        <Icon v-else icon="material-symbols:error" disabled="true" />
      </button>
      <button class="btn" :title="$t('Home.RssListButton.translate')" disabled>
        <Icon icon="material-symbols:g-translate" />
      </button>
      <button class="btn" @click="$emit('delete-history-content')" :title="$t('Home.RssListButton.deleteHistory')">
        <Icon icon="material-symbols:auto-delete-outline" />
      </button>
      <button class="btn" @click="goToSettings" :title="$t('Home.RssListButton.setting')">
        <Icon icon="material-symbols:settings" />
      </button>
    </div>
  </nav>
</template>

<style lang="scss" scoped>
@import '../../styles/home/RssListButton.scss';
</style>