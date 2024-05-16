<script setup lang="ts">
import { ref, watchEffect, onMounted } from 'vue';
import { Icon } from '@iconify/vue';
import { useI18n } from 'vue-i18n';

import { GetFeedList, SetFeedList, DeleteFeedList } from '../../../wailsjs/go/main/App'

const { t } = useI18n();

type FeedList = {
  Link: string
  Category: string
}

const feedList = ref([] as FeedList[])

async function getFeedList() {
  const result: FeedList[] = await GetFeedList()
  feedList.value = result
}

const selectedSubscribeType = ref('RSS/Atom');
const subscribeUrl = ref('');

async function setFeedList() {
  const feed: FeedList = {
    Link: subscribeUrl.value,
    Category: selectedSubscribeType.value
  }

  await SetFeedList([feed] as FeedList[])
  await getFeedList()

  selectedSubscribeType.value = 'RSS/Atom'
  subscribeUrl.value = ''
}

async function deleteFeedList(feed: FeedList) {
  await DeleteFeedList(feed)
  await getFeedList()
}

let subscribeUrlLabel = ref('');

watchEffect(() => {
  switch (selectedSubscribeType.value) {
    case 'RSS/Atom':
      subscribeUrlLabel.value = 'URL';
      break;
    case 'Twitter':
      subscribeUrlLabel.value = t('Settings.RssSettings.username');
      break;
    case 'Telegram':
      subscribeUrlLabel.value = 'ID';
      break;
    case 'Youtube':
      subscribeUrlLabel.value = t('Settings.RssSettings.username');
      break;
    case 'Wechat':
      subscribeUrlLabel.value = 'ID';
      break;
    default:
      subscribeUrlLabel.value = 'URL';
  }
});

onMounted(() => {
  getFeedList()
})
</script>

<template>
  <form name="new feed">
    <label for="subscribe-type">{{ $t('Settings.RssSettings.type') }}</label>
    <select id="subscribe-type" name="subscribe-type" v-model="selectedSubscribeType">
      <option value="RSS/Atom" selected>RSS/Atom</option>
      <option value="Twitter" disabled>Twitter</option>
      <option value="Telegram" disabled>Telegram</option>
      <option value="Youtube" disabled>Youtube</option>
      <option value="Wechat" disabled>{{ $t('Settings.RssSettings.wechat') }}</option>
    </select>
    <label for="subscribe-url">{{ subscribeUrlLabel }}</label>
    <input type="text" id="subscribe-url" name="subscribe-url" v-model="subscribeUrl" autocomplete="off"
      placeholder="https://feeds.bbci.co.uk/news/world/rss.xml" required />
    <button @click.prevent="setFeedList" class="btn" :title="$t('Settings.RssSettings.add')">
      <Icon icon="material-symbols:forms-add-on" />
    </button>
  </form>
  <ul v-if="feedList && feedList.length > 0">
    <li v-for="feed in feedList" :key="feed.Link">
      <div class="img">
        <img :src="`https://www.google.com/s2/favicons?domain=${feed.Link}`" alt="favicon" />
      </div>
      <span class="link">{{ feed.Link }}</span>
      <span class="category">{{ feed.Category }}</span>
      <button @click="deleteFeedList(feed)" class="btn" :title="$t('Settings.RssSettings.delete')">
        <Icon icon="material-symbols:delete-forever" />
      </button>
    </li>
  </ul>
  <div v-else class="NoFeedList">{{ $t('Settings.RssSettings.noFeedList') }}</div>
</template>

<style lang="scss" scoped>
@import '../../styles/settings/RssSettings.scss';
</style>