<template>
  <div class="flex items-center gap-3">
    <input
      type="checkbox"
      v-model="modelValue"
      title="Select All / Clear All"
      class="w-4 h-4 text-primary-600 rounded focus:ring-primary-500"
    >
    
    <button
      @click="$emit('deleteSelected')"
      title="Delete selected"
      class="p-2 rounded-md hover:bg-red-100 transition-colors"
    >
      <i class="fas fa-trash-alt"></i>
    </button>
    
    <button
      @click="$emit('copyFiltered')"
      title="Copy filtered"
      class="p-2 rounded-md hover:bg-gray-100 transition-colors"
    >
      <i class="fas fa-copy"></i>
    </button>
  </div>
</template>

<script setup>
const props = defineProps({
  modelValue: Boolean
});

const emit = defineEmits(['update:modelValue', 'deleteSelected', 'copyFiltered']);
</script>