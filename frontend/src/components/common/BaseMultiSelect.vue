<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onUnmounted, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhCheck } from '@phosphor-icons/vue';
import { type SelectOption, flattenOptions } from '@/types/select';
import { useSelect } from '@/composables/ui/useSelect';
import './select.css';

const { t } = useI18n();

// Generate unique ID for this dropdown instance
const dropdownId = `multiselect-${Math.random().toString(36).substring(2, 9)}`;

const props = defineProps({
  modelValue: {
    type: Array as PropType<(string | number)[]>,
    default: () => [],
  },
  options: {
    type: Array as PropType<SelectOption[]>,
    required: true,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  placeholder: {
    type: String,
    default: '',
  },
  width: {
    type: String,
    default: '',
  },
  maxWidth: {
    type: String,
    default: '',
  },
  maxHeight: {
    type: String,
    default: '',
  },
  searchable: {
    type: Boolean,
    default: false, // Reserved for future implementation
  },
  showSelectAll: {
    type: Boolean,
    default: true,
  },
  position: {
    type: String as PropType<'bottom' | 'top' | 'auto'>,
    default: 'bottom',
  },
  displayMode: {
    type: String as PropType<'chips' | 'counter'>,
    default: 'counter',
  },
});

const emit = defineEmits<{
  'update:modelValue': [value: (string | number)[]];
}>();

const isOpen = ref(false);
const dropdownPositionStyle = ref<Record<string, string>>({});

// Flatten options for navigation
const flatOptions = computed(() => flattenOptions(props.options));

// Use select composable for click outside and hover management
const {
  selectedIndex,
  triggerRef,
  dropdownRef,
  resetIndex,
  registerAsOpen,
  unregisterAsOpen,
  handleMouseLeave,
} = useSelect(flatOptions.value, isOpen, dropdownId);

// Compute dropdown position
const dropdownPosition = computed(() => {
  if (props.position === 'auto') {
    return 'bottom';
  }
  return props.position;
});

// Width class
const widthClass = computed(() => {
  if (props.width) {
    switch (props.width) {
      case 'sm':
        return 'w-20 sm:w-24';
      case 'md':
        return 'w-32 sm:w-48';
      case 'lg':
        return 'w-48 sm:w-64';
      default:
        return props.width;
    }
  }
  return 'w-full';
});

// Max width style
const maxWidthStyle = computed(() => {
  if (props.maxWidth) {
    return { maxWidth: props.maxWidth };
  }
  return {};
});

// Max height style for dropdown
const maxHeightStyle = computed(() => {
  if (props.maxHeight) {
    return { maxHeight: props.maxHeight };
  }
  return {};
});

// Calculate dropdown position when opened
function updateDropdownPosition() {
  if (!isOpen.value || !triggerRef.value) return;

  const triggerRect = triggerRef.value.getBoundingClientRect();
  const dropdownHeight = 200;
  const viewportHeight = window.innerHeight;

  let top: number;
  let position: 'top' | 'bottom' = dropdownPosition.value;

  if (props.position === 'auto') {
    const spaceBelow = viewportHeight - triggerRect.bottom;
    const spaceAbove = triggerRect.top;

    if (spaceBelow < dropdownHeight && spaceAbove > spaceBelow) {
      position = 'top';
    } else {
      position = 'bottom';
    }
  }

  if (position === 'bottom') {
    top = triggerRect.bottom + window.scrollY + 4;
  } else {
    top = triggerRect.top + window.scrollY - dropdownHeight - 4;
  }

  dropdownPositionStyle.value = {
    position: 'fixed',
    left: `${triggerRect.left}px`,
    top: `${top}px`,
    width: `${triggerRect.width}px`,
    zIndex: '9999',
  };
}

// Check if an option is selected
function isSelected(value: string | number): boolean {
  return props.modelValue.includes(value);
}

// Toggle an option
function toggleOption(option: SelectOption) {
  if (option.disabled) return;

  const newValues = isSelected(option.value)
    ? props.modelValue.filter((v) => v !== option.value)
    : [...props.modelValue, option.value];

  emit('update:modelValue', newValues);
}

// Select all
function selectAll() {
  const enabledOptions = flatOptions.value.filter((opt) => !opt.disabled);
  emit(
    'update:modelValue',
    enabledOptions.map((opt) => opt.value)
  );
}

// Deselect all
function deselectAll() {
  emit('update:modelValue', []);
}

// Check if all enabled options are selected
const allSelected = computed(() => {
  const enabledOptions = flatOptions.value.filter((opt) => !opt.disabled);
  if (enabledOptions.length === 0) return false;
  return enabledOptions.every((opt) => isSelected(opt.value));
});

// Check if some (but not all) enabled options are selected
const someSelected = computed(() => {
  const enabledOptions = flatOptions.value.filter((opt) => !opt.disabled);
  const selectedCount = enabledOptions.filter((opt) => isSelected(opt.value)).length;
  return selectedCount > 0 && selectedCount < enabledOptions.length;
});

// Display text for trigger button
const displayText = computed(() => {
  const count = props.modelValue.length;

  if (count === 0) {
    return props.placeholder || t('common.search.selectItems');
  }

  if (count === 1) {
    const option = flatOptions.value.find((opt) => opt.value === props.modelValue[0]);
    return option?.label || `${count} ${t('common.search.item')}`;
  }

  return t('common.search.itemsSelected', { count });
});

// Toggle dropdown
function toggleDropdown() {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    registerAsOpen();
    resetIndex();
    nextTick(updateDropdownPosition);
  } else {
    unregisterAsOpen();
  }
}

// Handle click on trigger
function handleTriggerClick(event: MouseEvent) {
  event.stopPropagation();
  toggleDropdown();
}

// Handle scroll to update position
function handleScroll() {
  if (isOpen.value) {
    updateDropdownPosition();
  }
}

// Handle window resize
function handleResize() {
  if (isOpen.value) {
    updateDropdownPosition();
  }
}

// Add/remove event listeners for scroll and resize
onMounted(() => {
  window.addEventListener('scroll', handleScroll, true);
  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll, true);
  window.removeEventListener('resize', handleResize);
});

// Get option class
function getOptionClass(option: SelectOption, index: number): string {
  const classes = ['select-option'];
  const isOptionSelected = isSelected(option.value);
  if (isOptionSelected) {
    classes.push('selected');
  }
  // Don't add focused class to selected items
  if (index === selectedIndex.value && isOpen.value && !isOptionSelected) {
    classes.push('focused');
  }
  if (option.disabled) {
    classes.push('disabled');
  }
  return classes.join(' ');
}

// Get trigger class
function getTriggerClass(): string {
  const classes = ['select-trigger'];
  if (props.disabled) {
    classes.push('disabled');
  }
  return classes.join(' ');
}

// Render selected chips
const selectedOptions = computed(() => {
  return props.modelValue
    .map((value) => flatOptions.value.find((opt) => opt.value === value))
    .filter(Boolean) as SelectOption[];
});
</script>

<template>
  <div :class="['select-container', 'relative', widthClass]" :style="maxWidthStyle">
    <!-- Display mode: chips -->
    <div v-if="displayMode === 'chips' && selectedOptions.length > 0" class="select-chips">
      <span
        v-for="option in selectedOptions"
        :key="option.value"
        class="select-chip"
        :style="{
          backgroundColor: option.color || 'var(--accent-color)',
          color: 'white',
        }"
      >
        {{ option.label }}
        <button type="button" class="hover:text-gray-200 ml-1" @click.stop="toggleOption(option)">
          ×
        </button>
      </span>
    </div>

    <!-- Trigger button -->
    <button
      ref="triggerRef"
      type="button"
      :class="getTriggerClass()"
      :disabled="disabled"
      :tabindex="disabled ? -1 : 0"
      @click="handleTriggerClick"
    >
      <span class="select-text truncate flex-1 text-left">
        {{ displayText }}
      </span>
      <PhCaretDown :size="14" :class="['select-arrow', { open: isOpen }]" />
    </button>

    <!-- Dropdown menu -->
    <Teleport to="body">
      <div
        v-if="isOpen"
        ref="dropdownRef"
        class="select-dropdown"
        :style="{ ...dropdownPositionStyle, ...maxHeightStyle }"
        @mouseleave="handleMouseLeave"
      >
        <!-- Select all / deselect all -->
        <div
          v-if="showSelectAll"
          class="select-all-container"
          @click="allSelected ? deselectAll() : selectAll()"
        >
          <input
            type="checkbox"
            :checked="allSelected"
            :indeterminate="someSelected"
            class="select-checkbox"
            tabindex="-1"
          />
          <span>
            {{ allSelected ? t('common.select.deselectAll') : t('common.select.selectAll') }}
          </span>
        </div>

        <!-- Options -->
        <div
          v-for="(option, index) in options"
          :key="option.value"
          :class="getOptionClass(option, index)"
          @click.stop="toggleOption(option)"
          @mouseenter="selectedIndex = index"
        >
          <input
            type="checkbox"
            :checked="isSelected(option.value)"
            :disabled="option.disabled"
            class="select-checkbox"
            tabindex="-1"
          />
          <span
            v-if="option.color"
            class="w-3 h-3 rounded-full flex-shrink-0"
            :style="{ backgroundColor: option.color }"
          />
          <span class="truncate flex-1">{{ option.label }}</span>
          <PhCheck
            v-if="isSelected(option.value)"
            :size="16"
            :class="['flex-shrink-0', isSelected(option.value) ? 'text-white' : 'text-accent']"
          />
        </div>

        <!-- Empty state -->
        <div v-if="flatOptions.length === 0" class="select-empty">
          {{ t('common.select.noOptions') }}
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.select-container {
  position: relative;
}

.select-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

input[type='checkbox']:indeterminate {
  appearance: none;
  width: 16px;
  height: 16px;
  border: 2px solid var(--accent-color);
  border-radius: 3px;
  background-color: var(--bg-primary);
  position: relative;
}

input[type='checkbox']:indeterminate::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 8px;
  height: 2px;
  background-color: var(--accent-color);
}
</style>
