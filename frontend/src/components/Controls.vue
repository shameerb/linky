<template>
  <div class="flex gap-3">
    <select
      v-model="modelValue"
      class="flex-1 p-2 border border-gray-300 rounded-md focus:border-primary-500 focus:ring-2 focus:ring-primary-200"
    >
      <option value="">Select a file</option>
      <option v-for="file in files" :key="file" :value="file">
        {{ file }}
      </option>
    </select>
    
    <input
      type="text"
      v-model="searchModelValue"
      placeholder="Search links..."
      class="flex-1 p-2 border border-gray-300 rounded-md focus:border-primary-500 focus:ring-2 focus:ring-primary-200"
    >
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  modelValue: String,
  searchQuery: String,
  files: Array
});

const emit = defineEmits(['update:modelValue', 'update:searchQuery']);

const modelValue = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
});

const searchModelValue = computed({
  get: () => props.searchQuery,
  set: (value) => emit('update:searchQuery', value)
});
</script>