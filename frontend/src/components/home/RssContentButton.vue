<script lang="ts" setup>
import { Icon } from '@iconify/vue';

const props = defineProps({
  selectedFeed: {
    type: Object,
    default: null
  }
})

const openInBrowser = () => {
  if (props.selectedFeed && props.selectedFeed.Link) {
    window.open(props.selectedFeed.Link, '_blank');
  }
}

const backToTop = () => {
  window.scrollTo(0, 0);
}
</script>

<template>
  <div class="ContentFunction">
    <button class="btn" @click="backToTop" :title="$t('Home.RssContentButton.backToTop')" disabled>
      <Icon icon="material-symbols:vertical-align-top" />
    </button>
    <button class="btn" :title="$t('Home.RssContentButton.bookmark')" disabled>
      <Icon icon="material-symbols:collections-bookmark" />
    </button>
    <button class="btn" @click="selectedFeed && $emit('modifyFeedContentReaded', selectedFeed, false)"
      :title="$t('Home.RssContentButton.unread')">
      <Icon icon="material-symbols:thread-unread" />
    </button>
    <button class="btn" @click="openInBrowser" :title="$t('Home.RssContentButton.openInBrowser')">
      <Icon icon="material-symbols:open-in-browser" />
    </button>
    <button class="btn" :title="$t('Home.RssContentButton.share')" disabled>
      <Icon icon="material-symbols:share" />
    </button>
    <button class="btn" :title="$t('Home.RssContentButton.translate')" disabled>
      <Icon icon="material-symbols:g-translate" />
    </button>
    <button class="btn" :title="$t('Home.RssContentButton.chatWithAI')" disabled>
      <Icon icon="material-symbols:robot-2" />
    </button>
  </div>
</template>

<style lang="scss" scoped>
@import '../../styles/home/RssContentButton.scss';
</style>