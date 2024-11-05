<!-- App.vue -->
<template>
  <div id="app">
    <div class="controls">
      <select v-model="selectedFile" @change="loadFile">
        <option value="">Select a file</option>
        <option v-for="file in files" :key="file" :value="file">{{ file }}</option>
      </select>
      <input type="text" v-model="searchQuery" @input="debouncedSearch" placeholder="Search links...">
    </div>
    
    <div id="preview-wrapper">
      <div class="multi-select-controls">
        <input type="checkbox" id="select-all-btn" v-model="selectAll" @change="toggleAllFiltered" title="Select All / Clear All">
        <button id="delete-selected-btn" @click="deleteSelected" title="Delete selected">
          <i class="fas fa-trash-alt"></i>
        </button>
        <button id="copy-filtered-btn" @click="copyFilteredLinks" title="Copy filtered">
          <i class="fas fa-copy"></i>
        </button>
      </div>
      <div id="preview">
        <div v-for="subject in filteredData" :key="subject.subject">
          <div class="subject">{{ subject.subject }}</div>
          <div v-for="link in subject.links" :key="link.url" class="link-item" :class="{ checked: link.checked }">
            <div class="checkbox-wrapper">
              <input type="checkbox" v-model="link.checked" @change="updateLinkStyle(link)">
            </div>
            <p>
              <a :href="link.url" target="_blank" :style="{ textDecoration: link.checked ? 'line-through' : 'none' }">
                {{ link.title }}
              </a>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { debounce } from 'lodash';

export default {
  name: 'App',
  data() {
    return {
      files: [],
      selectedFile: '',
      searchQuery: '',
      currentData: [],
      filteredData: [],
      selectAll: false,
    };
  },
  created() {
    this.loadFiles();
    document.addEventListener('keydown', this.handleKeyDown);
  },
  beforeDestroy() {
    document.removeEventListener('keydown', this.handleKeyDown);
  },
  methods: {
    async loadFiles() {
      try {
        const response = await fetch('/files');
        this.files = await response.json();
        
        const urlParams = new URLSearchParams(window.location.search);
        const fileParam = urlParams.get('file');
        const searchParam = urlParams.get('search');

        if (fileParam && this.files.includes(fileParam)) {
          this.selectedFile = fileParam;
          this.searchQuery = searchParam || '';
          await this.loadFile();
        } else if (this.files.length > 0) {
          this.selectedFile = this.files[0];
          await this.loadFile();
        }
      } catch (error) {
        console.error('Error loading files:', error);
      }
    },
    async loadFile() {
      try {
        const response = await fetch(`/file/${this.selectedFile}`);
        const data = await response.json();
        this.currentData = data.data;
        this.searchLinks();
        this.updateUrl();
      } catch (error) {
        console.error('Error loading file:', error);
      }
    },
    updateUrl() {
      const newUrl = new URL(window.location);
      newUrl.searchParams.set('file', this.selectedFile);
      if (this.searchQuery) {
        newUrl.searchParams.set('search', this.searchQuery);
      } else {
        newUrl.searchParams.delete('search');
      }
      window.history.pushState({}, '', newUrl);
    },
    searchLinks() {
      if (!this.searchQuery) {
        this.filteredData = this.currentData;
        return;
      }

      const keywords = this.searchQuery.toLowerCase().split(/\s+/).filter(keyword => keyword.length > 0);

      if (keywords.length === 0) {
        this.filteredData = this.currentData;
        return;
      }

      this.filteredData = this.currentData
        .filter(subject => subject.links && subject.links.length > 0)
        .map(subject => {
          const relevantLinks = subject.links.map(link => {
            const relevance = this.calculateRelevance(link, keywords);
            return relevance > 0 ? { ...link, relevance } : null;
          }).filter(Boolean);

          return relevantLinks.length > 0 ? { ...subject, links: relevantLinks } : null;
        }).filter(Boolean);

      // Sort links within each subject
      this.filteredData.forEach(subject => {
        subject.links.sort((a, b) => b.relevance - a.relevance);
      });

      // Sort subjects based on their most relevant link
      this.filteredData.sort((a, b) => b.links[0].relevance - a.links[0].relevance);
    },
    calculateRelevance(link, keywords) {
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
    },
    debouncedSearch: debounce(function() {
      this.searchLinks();
      this.updateUrl();
    }, 300),
    async deleteSelected() {
      const selectedLinks = this.getSelectedLinks();
      if (selectedLinks.length === 0) {
        alert('Please select at least one link to delete.');
        return;
      }
      
      if (confirm(`Are you sure you want to delete ${selectedLinks.length} selected link(s)?`)) {
        try {
          const response = await fetch('/delete_links', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              filename: this.selectedFile,
              links: selectedLinks
            }),
          });
          const data = await response.json();
          if (data.status === 'success') {
            await this.loadFile();
          } else {
            alert('Error deleting links: ' + data.message);
          }
        } catch (error) {
          console.error('Error deleting links:', error);
          alert('Error deleting links. Please try again.');
        }
      }
    },
    toggleAllFiltered(event) {
      const checked = event.target.checked;
      this.filteredData.forEach(subject => {
        subject.links.forEach(link => {
          link.checked = checked;
        });
      });
    },
    copyFilteredLinks() {
      const selectedLinks = this.getSelectedLinks();
      if (selectedLinks.length === 0) {
        alert('Please select at least one link to copy.');
        return;
      }

      const formattedLinks = selectedLinks.map(link => (`- [${link.title}](${link.url})`)).join('\n');
      
      navigator.clipboard.writeText(formattedLinks).then(() => {
        alert('Filtered links copied to clipboard!');
      }).catch(err => {
        console.error('Failed to copy links: ', err);
        alert('Failed to copy links. Please try again.');
      });
    },
    getSelectedLinks() {
      return this.filteredData.flatMap(subject => 
        subject.links.filter(link => link.checked)
      );
    },
    updateLinkStyle(link) {
      // This method is called when a checkbox is changed
      // Vue's reactivity will automatically update the UI
    },
    handleKeyDown(e) {
      if (e.ctrlKey && e.key === 'o') {
        e.preventDefault();
        this.$refs.fileSelect.focus();
      } else if (e.ctrlKey && e.key === 'f') {
        e.preventDefault();
        this.$refs.searchInput.focus();
      }
    }
  }
};
</script>

<style>
/* Include the styles from the original HTML file here */
:root {
  --primary-color: #3498db;
  --secondary-color: #2c3e50;
  --background-color: #f5f5f5;
  --card-background: #ffffff;
  --text-color: #333333;
  --border-color: #e0e0e0;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  max-width: 1000px;
  margin: 0 auto;
  background-color: #f5f5f5;
  color: #333;
  font-size: 14px;
  padding: 20px;
}

/* ... Include the rest of the styles from the original HTML ... */

</style>