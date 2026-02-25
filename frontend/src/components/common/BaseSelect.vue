<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onUnmounted, type PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { PhCaretDown, PhX } from '@phosphor-icons/vue';
import {
  type SelectOption,
  type SelectOptions,
  isOptionGroup,
  flattenOptions,
  findOption,
} from '@/types/select';
import { useSelect } from '@/composables/ui/useSelect';
import './select.css';

const { t } = useI18n();

// Generate unique ID for this dropdown instance
const dropdownId = `select-${Math.random().toString(36).substring(2, 9)}`;

const props = defineProps({
  modelValue: {
    type: [String, Number] as PropType<string | number>,
    required: true,
  },
  options: {
    type: Array as PropType<SelectOptions>,
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
  clearable: {
    type: Boolean,
    default: false,
  },
  position: {
    type: String as PropType<'bottom' | 'top' | 'auto'>,
    default: 'bottom',
  },
  size: {
    type: String as PropType<'xs' | 'sm' | 'md'>,
    default: 'sm',
  },
  allowCustomInput: {
    type: Boolean,
    default: false,
  },
  customInputPlaceholder: {
    type: String,
    default: '',
  },
  bgMode: {
    type: String as PropType<'primary' | 'secondary'>,
    default: 'primary',
  },
});

const emit = defineEmits<{
  'update:modelValue': [value: string | number];
  'custom-input': [value: string];
}>();

const isOpen = ref(false);
const customInputValue = ref('');
const customInputRef = ref<HTMLInputElement>();
const dropdownPositionStyle = ref<Record<string, string>>({});

// Flatten options for navigation
const flatOptions = computed(() => flattenOptions(props.options));

// Find current selected option
const selectedOption = computed(() => findOption(props.options, props.modelValue));

// Display text for trigger button
const displayText = computed(() => {
  if (selectedOption.value) {
    return selectedOption.value.label;
  }
  return props.placeholder || t('common.select.placeholder');
});

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
  const dropdownHeight = 200; // Approximate max height
  const viewportHeight = window.innerHeight;

  let top: number;
  let position: 'top' | 'bottom' = dropdownPosition.value;

  // Auto-detect position
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
    top = triggerRect.bottom + window.scrollY + 4; // 4px margin
  } else {
    top = triggerRect.top + window.scrollY - dropdownHeight - 4; // Approximate
  }

  dropdownPositionStyle.value = {
    position: 'fixed',
    left: `${triggerRect.left}px`,
    top: `${top}px`,
    width: `${triggerRect.width}px`,
    zIndex: '9999',
  };
}

// Toggle dropdown
function toggleDropdown() {
  if (props.disabled) return;
  isOpen.value = !isOpen.value;
  if (isOpen.value) {
    registerAsOpen();
    resetIndex();
    // Calculate position after DOM update
    nextTick(updateDropdownPosition);
  } else {
    unregisterAsOpen();
  }
}

// Select an option
function selectOption(option: SelectOption) {
  if (option.disabled) return;
  emit('update:modelValue', option.value);
  isOpen.value = false;
  unregisterAsOpen();
}

// Clear selection
function clearSelection(event: MouseEvent) {
  event.stopPropagation();
  emit('update:modelValue', '');
}

// Handle custom input
function handleCustomInput() {
  if (customInputValue.value.trim()) {
    emit('custom-input', customInputValue.value.trim());
    customInputValue.value = '';
    isOpen.value = false;
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

// Watch for model value changes from parent
watch(
  () => props.modelValue,
  () => {
    // Reset when value changes externally
  }
);

// Focus custom input when dropdown opens with allowCustomInput
watch(isOpen, (open) => {
  if (open && props.allowCustomInput) {
    nextTick(() => {
      customInputRef.value?.focus();
    });
  }
});

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
  const isSelected = option.value === props.modelValue;
  if (isSelected) {
    classes.push('selected');
  }
  // Don't add focused class to selected items
  if (index === selectedIndex.value && isOpen.value && !isSelected) {
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
  if (props.bgMode === 'secondary') {
    classes.push('select-mode-secondary');
  }
  if (!selectedOption.value && props.placeholder) {
    classes.push('select-placeholder');
  }
  return classes.join(' ');
}
</script>

<template>
  <div :class="['select-container', 'relative', widthClass]" :style="maxWidthStyle">
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
      <div class="flex items-center gap-1 flex-shrink-0">
        <PhX
          v-if="clearable && selectedOption && !disabled"
          :size="14"
          class="select-clear"
          @click.stop="clearSelection"
        />
        <PhCaretDown :size="14" :class="['select-arrow', { open: isOpen }]" />
      </div>
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
        <!-- Render grouped options -->
        <template v-for="(item, groupIndex) in options" :key="groupIndex">
          <!-- Option group -->
          <template v-if="isOptionGroup(item)">
            <div class="select-group-label">
              {{ item.label }}
            </div>
            <div
              v-for="option in item.options"
              :key="option.value"
              :class="getOptionClass(option, flatOptions.indexOf(option))"
              :style="option.style"
              @click.stop="selectOption(option)"
              @mouseenter="selectedIndex = flatOptions.indexOf(option)"
            >
              <slot name="option" :option="option">
                {{ option.label }}
              </slot>
            </div>
          </template>

          <!-- Single option -->
          <div
            v-else
            :class="getOptionClass(item, flatOptions.indexOf(item))"
            :style="item.style"
            @click.stop="selectOption(item)"
            @mouseenter="selectedIndex = flatOptions.indexOf(item)"
          >
            <slot name="option" :option="item">
              {{ item.label }}
            </slot>
          </div>
        </template>

        <!-- Custom input section -->
        <div v-if="allowCustomInput" class="select-custom-input">
          <input
            ref="customInputRef"
            v-model="customInputValue"
            type="text"
            :placeholder="customInputPlaceholder || t('common.input.customValue')"
            class="input-field w-full text-xs sm:text-sm"
            @keydown.enter="handleCustomInput"
          />
        </div>

        <!-- Empty state -->
        <div v-if="flatOptions.length === 0 && !allowCustomInput" class="select-empty">
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

.input-field {
  @apply p-1.5 sm:p-2 border border-border rounded-md bg-bg-secondary text-text-primary focus:border-accent focus:outline-none transition-colors text-xs sm:text-sm;
}
</style>
