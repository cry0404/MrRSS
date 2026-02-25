<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import BaseSelect from '@/components/common/BaseSelect.vue';
import type { SelectOption } from '@/types/select';

interface Props {
  category: string;
  categorySelection: string;
  showCustomCategory: boolean;
  existingCategories: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
  'update:category': [value: string];
  'update:categorySelection': [value: string];
  'update:showCustomCategory': [value: boolean];
  'handle-category-change': [value: string];
}>();

const { t } = useI18n();

function handleCategoryChange(value: string) {
  emit('handle-category-change', value);
}

// Build options for BaseSelect
const categoryOptions = computed<SelectOption[]>(() => {
  return [
    { value: '', label: t('sidebar.feedList.uncategorized') },
    ...props.existingCategories.map((cat) => ({
      value: cat,
      label: cat,
    })),
  ];
});

// Handle custom input from BaseSelect
function handleCustomInput(value: string) {
  emit('update:category', value);
  emit('update:categorySelection', value);
}

// Handle select value change
function handleSelectChange(value: string | number) {
  handleCategoryChange(String(value));
}
</script>

<template>
  <div class="mb-3 sm:mb-4">
    <label class="block mb-1 sm:mb-1.5 font-semibold text-xs sm:text-sm text-text-secondary">{{
      t('common.form.category')
    }}</label>

    <!-- Using BaseSelect with custom input support -->
    <BaseSelect
      v-if="!props.showCustomCategory"
      :model-value="props.categorySelection"
      :options="categoryOptions"
      :allow-custom-input="true"
      :custom-input-placeholder="t('modal.feed.enterCategoryName')"
      :searchable="true"
      @update:model-value="handleSelectChange"
      @custom-input="handleCustomInput"
    />

    <!-- Custom category input (legacy mode for backward compatibility) -->
    <div v-else class="flex gap-2">
      <input
        :value="props.category"
        type="text"
        :placeholder="t('modal.feed.enterCategoryName')"
        class="input-field flex-1"
        autofocus
        @input="emit('update:category', ($event.target as HTMLInputElement).value)"
      />
      <button
        type="button"
        class="px-3 py-2 text-xs sm:text-sm text-text-secondary hover:text-text-primary border border-border rounded-md hover:bg-bg-tertiary transition-colors"
        @click="
          emit('update:showCustomCategory', false);
          emit('update:categorySelection', '');
        "
      >
        {{ t('common.cancel') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
@reference "../../../style.css";

.input-field {
  @apply w-full p-2 sm:p-2.5 border border-border rounded-md bg-bg-tertiary text-text-primary text-xs sm:text-sm focus:border-accent focus:outline-none transition-colors;
  box-sizing: border-box;
}
</style>
