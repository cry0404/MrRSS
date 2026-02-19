<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhArrowLineUp } from '@phosphor-icons/vue';

interface TocItem {
  id: string;
  text: string;
  level: 1 | 2 | 3;
  offsetTop: number;
  markerWidth: number;
  parentIndex: number | null;
  isFallback: boolean;
}

interface Props {
  articleId: number;
  enabled: boolean;
  scrollContainer: HTMLElement | null;
}

const props = defineProps<Props>();
const { t } = useI18n();

const tocItems = ref<TocItem[]>([]);
const activeIndex = ref(-1);
const sectionProgress = ref(0);
const articleProgress = ref(0);
const isDesktop = ref(false);
const topOffset = ref(76);
const bottomOffset = ref(24);

let mediaQuery: MediaQueryList | null = null;
let containerObserver: MutationObserver | null = null;
let pendingProseObserver: MutationObserver | null = null;
let uiObserver: MutationObserver | null = null;
let scrollContainerEl: HTMLElement | null = null;
let rebuildRaf: number | null = null;
let scrollRaf: number | null = null;

const containerStyle = computed(() => ({
  top: `${topOffset.value}px`,
  bottom: `${bottomOffset.value}px`,
}));

function getMarkerWidth(level: 1 | 2 | 3): number {
  if (level === 1) return 34;
  if (level === 2) return 24;
  return 14;
}

function sanitizeHeadingText(text: string): string {
  // Keep numbered prefixes like "1.2.", but remove markdown heading markers like "###".
  return text
    .replace(/^\s*#+\s*/, '')
    .replace(/\s+/g, ' ')
    .trim();
}

function getHeadingDisplayText(heading: HTMLElement): string {
  const translationEl = heading.nextElementSibling as HTMLElement | null;
  const hasHeadingTranslation =
    translationEl &&
    translationEl.classList.contains('translation-text') &&
    !translationEl.classList.contains('translation-inline') &&
    !translationEl.classList.contains('translation-blockquote');

  const translatedText = hasHeadingTranslation
    ? sanitizeHeadingText(translationEl.textContent || '')
    : '';
  if (translatedText) return translatedText;

  return sanitizeHeadingText(heading.textContent || '');
}

function shouldShowText(itemIndex: number): boolean {
  const item = tocItems.value[itemIndex];
  if (!item || item.isFallback) {
    return false;
  }

  if (activeIndex.value < 0 || activeIndex.value >= tocItems.value.length) {
    return false;
  }
  if (itemIndex === activeIndex.value) {
    return true;
  }

  let parent = tocItems.value[activeIndex.value].parentIndex;
  while (parent !== null) {
    if (parent === itemIndex) {
      return true;
    }
    parent = tocItems.value[parent].parentIndex;
  }
  return false;
}

function queueRebuild(): void {
  if (rebuildRaf !== null) {
    cancelAnimationFrame(rebuildRaf);
  }
  rebuildRaf = requestAnimationFrame(() => {
    rebuildRaf = null;
    buildToc();
    updateAvoidanceOffsets();
  });
}

function queueScrollSync(): void {
  if (scrollRaf !== null) return;
  scrollRaf = requestAnimationFrame(() => {
    scrollRaf = null;
    updateActiveSection();
  });
}

function getArticleProgress(container: HTMLElement): number {
  const maxScrollTop = Math.max(0, container.scrollHeight - container.clientHeight);
  if (maxScrollTop <= 0) return 100;
  const progress = Math.max(0, Math.min(1, container.scrollTop / maxScrollTop));
  return Math.round(progress * 100);
}

function getMarkerFillPercent(index: number): number {
  if (index !== activeIndex.value) return 0;
  return sectionProgress.value;
}

function buildToc(): void {
  const container = props.scrollContainer;
  if (!props.enabled || !isDesktop.value || !container) {
    tocItems.value = [];
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = 0;
    return;
  }

  const proseContainer = container.querySelector('.prose-content');
  if (!proseContainer) {
    tocItems.value = [];
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = getArticleProgress(container);
    return;
  }

  const headings = Array.from(proseContainer.querySelectorAll<HTMLElement>('h1, h2, h3'));
  const containerRect = container.getBoundingClientRect();
  const articleId = props.articleId || 0;

  const items: TocItem[] = [];
  let lastH1Index: number | null = null;
  let lastH2Index: number | null = null;
  headings.forEach((heading, index) => {
    const level = Number(heading.tagName.slice(1)) as 1 | 2 | 3;
    if (level < 1 || level > 3) return;

    const text = getHeadingDisplayText(heading);
    if (!text) return;

    if (!heading.id) {
      heading.id = `toc-heading-${articleId}-${index}`;
    }

    const offsetTop = heading.getBoundingClientRect().top - containerRect.top + container.scrollTop;

    let parentIndex: number | null = null;
    if (level === 2) {
      parentIndex = lastH1Index;
    } else if (level === 3) {
      parentIndex = lastH2Index ?? lastH1Index;
    }

    items.push({
      id: heading.id,
      text,
      level,
      offsetTop: Math.max(0, Math.round(offsetTop)),
      markerWidth: getMarkerWidth(level),
      parentIndex,
      isFallback: false,
    });

    const itemIndex = items.length - 1;
    if (level === 1) {
      lastH1Index = itemIndex;
      lastH2Index = null;
    } else if (level === 2) {
      lastH2Index = itemIndex;
    }
  });

  if (items.length === 0) {
    items.push({
      id: `toc-fallback-${articleId}`,
      text: '',
      level: 1,
      offsetTop: 0,
      markerWidth: getMarkerWidth(1),
      parentIndex: null,
      isFallback: true,
    });
  }

  tocItems.value = items;
  updateActiveSection();
}

function updateActiveSection(): void {
  const container = props.scrollContainer;
  const items = tocItems.value;

  if (!container || items.length === 0) {
    activeIndex.value = -1;
    sectionProgress.value = 0;
    articleProgress.value = container ? getArticleProgress(container) : 0;
    return;
  }

  articleProgress.value = getArticleProgress(container);

  const pointer = container.scrollTop + 12;
  let currentIndex = 0;

  for (let i = 0; i < items.length; i += 1) {
    if (items[i].offsetTop <= pointer) {
      currentIndex = i;
    } else {
      break;
    }
  }

  activeIndex.value = currentIndex;

  const start = items[currentIndex].offsetTop;
  const maxScrollTop = Math.max(0, container.scrollHeight - container.clientHeight);
  const end =
    currentIndex < items.length - 1
      ? items[currentIndex + 1].offsetTop
      : Math.max(start + 1, maxScrollTop);

  const segmentSize = Math.max(1, end - start);
  const progress = Math.max(0, Math.min(1, (pointer - start) / segmentSize));
  sectionProgress.value = Math.round(progress * 100);
}

function scrollToHeading(item: TocItem): void {
  const container = props.scrollContainer;
  if (!container) return;

  const targetTop = Math.max(0, item.offsetTop - 12);
  container.scrollTo({
    top: targetTop,
    behavior: 'smooth',
  });
}

function scrollToTop(): void {
  const container = props.scrollContainer;
  if (!container) return;

  container.scrollTo({
    top: 0,
    behavior: 'smooth',
  });
}

function updateAvoidanceOffsets(): void {
  if (!isDesktop.value) return;

  let nextTop = 76;
  let nextBottom = 24;

  const findBar = document.querySelector('.find-in-page-bar') as HTMLElement | null;
  if (findBar) {
    const rect = findBar.getBoundingClientRect();
    nextTop = Math.max(nextTop, Math.round(rect.bottom + 12));
  }

  const chatPanel = document.querySelector('.chat-panel') as HTMLElement | null;
  if (chatPanel) {
    const rect = chatPanel.getBoundingClientRect();
    nextBottom = Math.max(nextBottom, Math.round(window.innerHeight - rect.top + 12));
  } else {
    const chatButton = document.querySelector('.js-article-chat-button') as HTMLElement | null;
    if (chatButton) {
      const rect = chatButton.getBoundingClientRect();
      nextBottom = Math.max(nextBottom, Math.round(window.innerHeight - rect.top + 12));
    }
  }

  topOffset.value = nextTop;
  bottomOffset.value = nextBottom;
}

function handleMediaChange(event: MediaQueryListEvent): void {
  isDesktop.value = event.matches;
  queueRebuild();
}

function bindScrollContainer(container: HTMLElement | null): void {
  if (scrollContainerEl) {
    scrollContainerEl.removeEventListener('scroll', queueScrollSync);
  }

  scrollContainerEl = container;

  if (scrollContainerEl) {
    scrollContainerEl.addEventListener('scroll', queueScrollSync, { passive: true });
  }
}

function connectContainerObserver(): void {
  containerObserver?.disconnect();
  pendingProseObserver?.disconnect();
  containerObserver = null;
  pendingProseObserver = null;

  const container = props.scrollContainer;
  if (!container) return;

  const proseContainer = container.querySelector('.prose-content');
  if (!proseContainer) {
    // The article body may render asynchronously. Watch container until prose appears.
    pendingProseObserver = new MutationObserver(() => {
      const readyProse = container.querySelector('.prose-content');
      if (!readyProse) return;

      pendingProseObserver?.disconnect();
      pendingProseObserver = null;
      connectContainerObserver();
      queueRebuild();
    });

    pendingProseObserver.observe(container, {
      childList: true,
      subtree: true,
    });
    return;
  }

  containerObserver = new MutationObserver(() => queueRebuild());
  containerObserver.observe(proseContainer, {
    childList: true,
    subtree: true,
    characterData: true,
  });
}

onMounted(async () => {
  mediaQuery = window.matchMedia('(min-width: 768px)');
  isDesktop.value = mediaQuery.matches;

  mediaQuery.addEventListener('change', handleMediaChange);
  window.addEventListener('resize', queueRebuild);

  // Watch floating UI visibility changes (find bar/chat panel) for dynamic avoidance.
  uiObserver = new MutationObserver(() => updateAvoidanceOffsets());
  uiObserver.observe(document.body, {
    childList: true,
    subtree: true,
  });

  await nextTick();
  bindScrollContainer(props.scrollContainer);
  connectContainerObserver();
  queueRebuild();
});

watch(
  () => [props.articleId, props.enabled] as const,
  async () => {
    await nextTick();
    bindScrollContainer(props.scrollContainer);
    connectContainerObserver();
    queueRebuild();
  }
);

watch(
  () => props.scrollContainer,
  async (container) => {
    await nextTick();
    bindScrollContainer(container);
    connectContainerObserver();
    queueRebuild();
  }
);

onBeforeUnmount(() => {
  if (mediaQuery) {
    mediaQuery.removeEventListener('change', handleMediaChange);
  }

  window.removeEventListener('resize', queueRebuild);
  scrollContainerEl?.removeEventListener('scroll', queueScrollSync);
  containerObserver?.disconnect();
  pendingProseObserver?.disconnect();
  uiObserver?.disconnect();

  if (rebuildRaf !== null) {
    cancelAnimationFrame(rebuildRaf);
  }
  if (scrollRaf !== null) {
    cancelAnimationFrame(scrollRaf);
  }
});
</script>

<template>
  <div v-if="enabled && isDesktop && tocItems.length > 0"
    class="fixed right-4 z-40 flex [width:max(15%,125px)] flex-col items-end justify-center [container-type:inline-size]"
    :style="containerStyle">
    <div class="mb-2 w-full text-right text-[10px] font-medium text-text-secondary opacity-75">
      {{ articleProgress }}%
    </div>

    <div class="group/toclist relative w-full max-h-[80%]">
      <div
        class="pointer-events-none absolute -inset-y-1.5 -left-2 -right-1 rounded-lg border border-border bg-bg-secondary shadow-lg shadow-black/15 opacity-0 scale-[0.98] transition-all duration-200 group-hover/toclist:opacity-100 group-hover/toclist:scale-100 dark:shadow-black/40">
      </div>

      <ul
        class="relative z-[1] flex w-full max-h-full flex-col items-end gap-1 overflow-hidden group-hover/toclist:overflow-auto">
        <li v-for="(item, index) in tocItems" :key="item.id" class="w-full" :data-level="item.level">
          <button
            class="group/item flex w-full cursor-pointer items-center justify-end gap-2 rounded px-1 py-0.5 transition-colors"
            :style="{ '--toc-level': String(item.level) }" @click="scrollToHeading(item)">
            <span :class="[
              'toc-text truncate text-right text-xs opacity-0 transition-all duration-200 [margin-left:calc((var(--toc-level,1)-1)*12px)]',
              index === activeIndex ? 'text-text-primary' : 'text-text-secondary',
              shouldShowText(index)
                ? 'toc-text-visible opacity-[0.85] max-w-[calc(100%-42px)]'
                : 'max-w-0',
              'group-hover/toclist:opacity-[0.85] group-hover/toclist:max-w-[calc(100%-42px)] group-hover/item:text-text-primary group-hover/item:opacity-100',
            ]" :data-level="item.level" :title="item.text">
              {{ item.text }}
            </span>
            <span class="flex w-[34px] justify-end">
              <span :class="[
                'relative overflow-hidden bg-text-secondary transition-colors duration-150 group-hover/item:bg-text-primary',
                index === activeIndex ? 'h-[3px] opacity-100' : 'h-[2px] opacity-70',
              ]" :style="{ width: `${item.markerWidth}px` }">
                <span
                  class="absolute left-0 top-0 h-full bg-accent transition-all duration-150 group-hover/item:bg-text-primary"
                  :style="{ width: `${getMarkerFillPercent(index)}%` }"></span>
              </span>
            </span>
          </button>
        </li>
      </ul>
    </div>

    <button
      class="mt-3 flex h-7 w-7 items-center justify-center self-end rounded bg-transparent text-text-secondary transition-colors hover:bg-[color-mix(in_srgb,var(--bg-tertiary)_70%,transparent)] hover:text-text-primary"
      :title="t('common.back')" @click="scrollToTop">
      <PhArrowLineUp :size="14" />
    </button>
  </div>

  <template v-else-if="enabled && !isDesktop">
    <!-- TODO: Add floating TOC UI for mobile devices. -->
  </template>
</template>

<style scoped>
@container (max-width: 150px) {
  .toc-text-visible {
    opacity: 0;
    max-width: 0;
  }
}
</style>
