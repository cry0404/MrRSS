import { ref, onMounted, onUnmounted, type Ref } from 'vue';
import type { SelectOption } from '@/types/select';

// Global state to track currently open dropdown
let openDropdownId: string | null = null;
let registerOpenDropdown: ((id: string | null) => void) | null = null;

export function useSelect(options: SelectOption[], isOpen: Ref<boolean>, dropdownId?: string) {
  const selectedIndex = ref(-1);
  const triggerRef = ref<HTMLElement>();
  const dropdownRef = ref<HTMLElement>();

  // Reset selected index
  function resetIndex() {
    selectedIndex.value = -1;
  }

  // Register this dropdown as the open one
  function registerAsOpen() {
    if (dropdownId) {
      openDropdownId = dropdownId;
      // Notify other dropdowns to close
      if (registerOpenDropdown) {
        registerOpenDropdown(dropdownId);
      }
    }
  }

  // Unregister this dropdown
  function unregisterAsOpen() {
    if (dropdownId && openDropdownId === dropdownId) {
      openDropdownId = null;
      if (registerOpenDropdown) {
        registerOpenDropdown(null);
      }
    }
  }

  // Handle other dropdown opening
  function onOtherDropdownOpen(id: string | null) {
    if (dropdownId && id !== dropdownId) {
      isOpen.value = false;
      resetIndex();
    }
  }

  // Click outside to close
  function handleClickOutside(event: MouseEvent) {
    if (
      isOpen.value &&
      triggerRef.value &&
      dropdownRef.value &&
      !triggerRef.value.contains(event.target as Node) &&
      !dropdownRef.value.contains(event.target as Node)
    ) {
      isOpen.value = false;
      resetIndex();
      unregisterAsOpen();
    }
  }

  // Mouse leaves dropdown area - reset hover state
  function handleMouseLeave() {
    if (isOpen.value) {
      selectedIndex.value = -1;
    }
  }

  onMounted(() => {
    document.addEventListener('click', handleClickOutside);
    // Register for other dropdown notifications
    if (dropdownId && !registerOpenDropdown) {
      registerOpenDropdown = onOtherDropdownOpen;
    }
  });

  onUnmounted(() => {
    document.removeEventListener('click', handleClickOutside);
    unregisterAsOpen();
    if (registerOpenDropdown === onOtherDropdownOpen) {
      registerOpenDropdown = null;
    }
  });

  return {
    selectedIndex,
    triggerRef,
    dropdownRef,
    resetIndex,
    registerAsOpen,
    unregisterAsOpen,
    handleMouseLeave,
  };
}
