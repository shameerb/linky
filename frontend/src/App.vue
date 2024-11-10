// App.vue
<template>
  <div class="max-w-4xl mx-auto p-5 bg-gray-50">
    <Controls 
      v-model:selectedFile="selectedFile"
      v-model:searchQuery="searchQuery"
      :files="files"
    />
    
    <div class="bg-white border border-gray-200 rounded-lg shadow-sm p-4 mt-4">
      <div class="flex justify-between items-center mb-4">
        <MultiSelectControls
          v-model:selectAll="selectAll"
          @delete-selected="deleteSelected"
          @copy-filtered="copyFilteredLinks"
        />
        <div class="text-sm text-gray-600">
          {{ totalLinks }} links
        </div>
      </div>
      
      <LinksList
        :data="filteredData"
        v-model:selectedLinks="selectedLinks"
        :selectAll="selectAll"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue';
import Controls from './components/Controls.vue';
import MultiSelectControls from './components/MultiSelectControls.vue';
import LinksList from './components/LinksList.vue';

const files = ref([]);
const selectedFile = ref('');
const searchQuery = ref('');
const currentData = ref([]);
const selectedLinks = ref(new Set());
const selectAll = ref(false);

const totalLinks = computed(() => {
  return currentData.value.reduce((total, subject) => 
    total + (subject.links?.length || 0), 0);
});

const filteredData = computed(() => {
  if (!searchQuery.value) return currentData.value;
  
  const keywords = searchQuery.value.toLowerCase().split(/\s+/).filter(k => k.length > 0);
  if (keywords.length === 0) return currentData.value;

  return currentData.value
    .filter(subject => subject.links?.length > 0)
    .map(subject => {
      const relevantLinks = subject.links
        .map(link => {
          const relevance = calculateRelevance(link, keywords);
          return relevance > 0 ? { ...link, relevance } : null;
        })
        .filter(Boolean)
        .sort((a, b) => b.relevance - a.relevance);

      return relevantLinks.length > 0 ? { ...subject, links: relevantLinks } : null;
    })
    .filter(Boolean)
    .sort((a, b) => b.links[0].relevance - a.links[0].relevance);
});

function calculateRelevance(link, keywords) {
  const titleWords = link.title.toLowerCase().split(/\s+/);
  const urlWords = link.url.toLowerCase().split(/\s+/);
  
  let relevance = 0;
  let matchedKeywords = 0;

  for (const keyword of keywords) {
    let keywordMatched = false;

    for (const word of titleWords) {
      if (word === keyword) {
        relevance += 2;
        keywordMatched = true;
        break;
      } else if (word.includes(keyword)) {
        relevance += 1;
        keywordMatched = true;
        break;
      }
    }

    if (!keywordMatched) {
      for (const word of urlWords) {
        if (word === keyword) {
          relevance += 1;
          keywordMatched = true;
          break;
        } else if (word.includes(keyword)) {
          relevance += 0.5;
          keywordMatched = true;
          break;
        }
      }
    }

    if (keywordMatched) matchedKeywords++;
  }

  return matchedKeywords === keywords.length ? relevance : 0;
}

async function loadFiles() {
  const response = await fetch('/files');
  files.value = await response.json();
  
  const urlParams = new URLSearchParams(window.location.search);
  const fileParam = urlParams.get('file');
  const searchParam = urlParams.get('search');

  if (fileParam && files.value.includes(fileParam)) {
    selectedFile.value = fileParam;
    searchQuery.value = searchParam || '';
    await loadFile(fileParam);
  } else if (files.value.length > 0) {
    selectedFile.value = files.value[0];
    await loadFile(files.value[0]);
  }
}

async function loadFile(filename) {
  const response = await fetch(`/file/${filename}`);
  const result = await response.json();
  currentData.value = result.data;
  updateUrl(filename, searchQuery.value);
}

function updateUrl(filename, search = '') {
  const newUrl = new URL(window.location);
  newUrl.searchParams.set('file', filename);
  if (search) {
    newUrl.searchParams.set('search', search);
  } else {
    newUrl.searchParams.delete('search');
  }
  window.history.pushState({}, '', newUrl);
}

async function deleteSelected() {
  if (selectedLinks.value.size === 0) {
    alert('Please select at least one link to delete.');
    return;
  }

  if (!confirm(`Are you sure you want to delete ${selectedLinks.value.size} selected link(s)?`)) {
    return;
  }

  const linksToDelete = Array.from(selectedLinks.value).map(link => ({
    title: link.title,
    url: link.url
  }));

  const response = await fetch('/delete_links', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      filename: selectedFile.value,
      links: linksToDelete
    })
  });

  const result = await response.json();
  if (result.status === 'success') {
    await loadFile(selectedFile.value);
    selectedLinks.value.clear();
    selectAll.value = false;
  } else {
    alert('Error deleting links: ' + result.message);
  }
}

function copyFilteredLinks() {
  if (selectedLinks.value.size === 0) {
    alert('Please select at least one link to copy.');
    return;
  }

  const formattedLinks = Array.from(selectedLinks.value)
    .map(link => `- [${link.title}](${link.url})`)
    .join('\n');

  navigator.clipboard.writeText(formattedLinks)
    .then(() => alert('Filtered links copied to clipboard!'))
    .catch(() => alert('Failed to copy links. Please try again.'));
}

watch(selectedFile, loadFile);

watch(searchQuery, (value) => {
  updateUrl(selectedFile.value, value);
});

onMounted(() => {
  loadFiles();
  
  window.addEventListener('keydown', (e) => {
    if (e.ctrlKey && e.key === 'o') {
      e.preventDefault();
      document.querySelector('select').focus();
    } else if (e.ctrlKey && e.key === 'f') {
      e.preventDefault();
      document.querySelector('input[type="text"]').focus();
    }
  });
});
</script>

<style>
:root {
  --primary-color: #3498db;
  --secondary-color: #2c3e50;
  --background-color: #f5f5f5;
  --card-background: #ffffff;
  --text-color: #333333;
  --border-color: #e0e0e0;
}

.link-item.checked a {
  text-decoration: line-through;
}
</style>