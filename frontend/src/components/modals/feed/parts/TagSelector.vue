<script setup lang="ts">
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAppStore } from '@/stores/app';
import { PhX, PhPlus } from '@phosphor-icons/vue';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption } from '@/types/select';
import TagFormModal from '../../settings/tags/TagFormModal.vue';

interface Props {
  selectedTags: number[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:selectedTags': [value: number[]];
}>();

const { t } = useI18n();
const store = useAppStore();

const availableTags = computed(() => store.tags || []);

// Build options for BaseSelect
const tagOptions = computed<SelectOption[]>(() => {
  return [
    { value: '', label: t('modal.tag.selectTags') },
    ...availableTags.value.map((tag) => ({
      value: tag.id,
      label: tag.name,
      disabled: props.selectedTags.includes(tag.id),
    })),
  ];
});

// New tag creation state
const showNewTagModal = ref(false);

function toggleTag(tagId: number) {
  const newSelection = props.selectedTags.includes(tagId)
    ? props.selectedTags.filter((id) => id !== tagId)
    : [...props.selectedTags, tagId];
  emit('update:selectedTags', newSelection);
}

function removeTag(tagId: number) {
  emit(
    'update:selectedTags',
    props.selectedTags.filter((id) => id !== tagId)
  );
}

function openNewTagModal() {
  showNewTagModal.value = true;
}

function closeNewTagModal() {
  showNewTagModal.value = false;
}

async function handleSaveTag(name: string, color: string) {
  try {
    const res = await fetch('/api/tags', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, color }),
    });

    if (res.ok) {
      const newTag = await res.json();
      // Refresh tags in store
      await store.fetchTags();
      // Auto-select the newly created tag
      emit('update:selectedTags', [...props.selectedTags, newTag.id]);
      closeNewTagModal();
      window.showToast(t('modal.tag.tagCreated'), 'success');
    } else {
      window.showToast(t('common.errors.createFailed'), 'error');
    }
  } catch (e) {
    console.error('Failed to create tag:', e);
    window.showToast(t('common.errors.createFailed'), 'error');
  }
}

// Handle select change
function handleSelectChange(value: string | number) {
  if (value !== '') {
    toggleTag(Number(value));
  }
}
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">{{
      t('modal.tag.selectTags')
    }}</label>

    <!-- Selected tags as chips -->
    <div v-if="selectedTags.length > 0" class="flex flex-wrap gap-2 mb-2">
      <span
        v-for="tagId in selectedTags"
        :key="tagId"
        class="inline-flex items-center gap-1 px-2 py-1 text-xs rounded text-white"
        :style="{ backgroundColor: store.tagMap.get(tagId)?.color || '#3B82F6' }"
      >
        {{ store.tagMap.get(tagId)?.name }}
        <button class="hover:text-gray-200" @click="removeTag(tagId)">
          <PhX :size="14" />
        </button>
      </span>
    </div>

    <!-- Available tags dropdown -->
    <BaseSelect
      :model-value="''"
      :options="tagOptions"
      :searchable="true"
      @update:model-value="handleSelectChange"
    />

    <!-- Create new tag button -->
    <button
      type="button"
      class="mt-2 text-xs text-accent hover:text-accent-hover transition-colors flex items-center gap-1"
      @click="openNewTagModal"
    >
      <PhPlus :size="14" />
      {{ t('modal.tag.createNew') }}
    </button>

    <!-- New tag modal -->
    <Teleport to="body">
      <TagFormModal
        v-if="showNewTagModal"
        :editing-tag="null"
        @close="closeNewTagModal"
        @save="handleSaveTag"
      />
    </Teleport>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

/* Styles are now handled by BaseSelect and select.css */
</style>
