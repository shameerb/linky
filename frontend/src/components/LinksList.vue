<template>
  <div>
    <template v-for="subject in data" :key="subject.subject">
      <div class="text-lg font-semibold text-gray-700 mt-5 mb-3">
        {{ subject.subject }}
      </div>
      
      <div
        v-for="link in subject.links"
        :key="link.url"
        :class="['flex items-center p-2 border-b border-gray-100 last:border-0', 
                 { 'checked': isSelected(link) }]"
      >
        <div class="w-6">
          <input
            type="checkbox"
            :checked="isSelected(link)"
            @change="toggleLink(link)"
            class="w-4 h-4 text-primary-600 rounded focus:ring-primary-500"
          >
        </div>
        
        <p class="m-0 flex-grow">
          <a
            :href="link.url"
            target="_blank"
            class="text-blue-900 hover:underline"
            :class="{ 'line-through': isSelected(link) }"
          >
            {{ link.title }}
          </a>
        </p>
      </div>
    </template>
  </div>
</template>

<script setup>
import { watch } from 'vue';

const props = defineProps({
  data: Array,
  selectedLinks: Set,
  selectAll: Boolean
});

const emit = defineEmits(['update:selectedLinks']);

function isSelected(link) {
  return props.selectedLinks.has(link);
}

function toggleLink(link) {
  const newSet = new Set(props.selectedLinks);
  if (newSet.has(link)) {
    newSet.delete(link);
  } else {
    newSet.add(link);
  }
  emit('update:selectedLinks', newSet);
}

watch(() => props.selectAll, (newValue) => {
  const newSet = new Set();
  if (newValue) {
    props.data.forEach(subject => {
      subject.links.forEach(link => newSet.add(link));
    });
  }
  emit('update:selectedLinks', newSet);
});
</script>