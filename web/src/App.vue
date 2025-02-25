<!-- App.vue -->
<template>
  <div class="app">
    <div class="controls">
      <select v-model="currentFile" id="file-list">
        <option value="">Select a file</option>
        <option v-for="file in files" :key="file" :value="file">
          {{ file }}
        </option>
      </select>
      <input
        type="text"
        v-model="searchQuery"
        @input="onSearchInput"
        placeholder="Search links..."
        id="search-input"
      />
    </div>

    <div id="preview-wrapper">
      <div class="main-content">
        <div class="multi-select-controls">
          <div class="control-buttons">
            <input 
              type="checkbox" 
              id="select-all-btn" 
              title="Select All / Clear All"
              @change="toggleAllFiltered"
              :checked="allSelected"
            >
            <button 
              id="delete-selected-btn" 
              title="Delete selected"
              @click="deleteSelected"
            >
              <font-awesome-icon icon="trash-alt" />
            </button>
            <button 
              id="copy-filtered-btn" 
              title="Copy filtered"
              @click="copyFilteredLinks"
            >
              <font-awesome-icon icon="copy" />
            </button>
            <button 
              id="bulk-add-btn" 
              title="Add bulk links to new section"
              @click="openBulkAddDialog('')"
            >
              <font-awesome-icon icon="plus" />
            </button>
          </div>
          <span class="count-info">
            Total: <span id="total-links">{{ totalLinks }}</span>
          </span>
        </div>

        <div v-for="subject in filteredSubjects" :key="subject.subject" class="subject-group">
          <div class="subject-header">
            <div class="subject">{{ subject.subject }}</div>
          </div>
          <div class="links-list">
            <div 
              v-for="(link, index) in subject.links" 
              :key="link.url" 
              class="link-item"
              :class="{ selected: link.selected, focused: focusedLinkIndex === index }"
              :tabindex="0"
              :data-url="link.url"
              @focus="focusedLinkIndex = index"
              @blur="focusedLinkIndex = -1"
            >
              <div class="checkbox-wrapper">
                <input 
                  type="checkbox" 
                  v-model="link.selected" 
                  @change="onCheckboxChange(link)"
                >
              </div>
              <p>
                <a :href="link.url" target="_blank" @click.prevent="openLink(link)">{{ link.title }}</a>
              </p>
            </div>
          </div>
        </div>

        <div v-if="!hasLinks" class="no-links">
          No links found
        </div>
      </div>
    </div>

    <BulkLinkAdder
      v-if="showBulkDialog"
      :current-file="currentFile"
      :available-subjects="subjects.map(s => s.subject)"
      @links-added="handleLinksAdded"
      @close="showBulkDialog = false"
    />

    <!-- Help button -->
    <button 
      class="help-button" 
      @click="showHelp = true"
      title="Keyboard Shortcuts"
    >?</button>

    <!-- Help dialog -->
    <KeyboardShortcuts 
      :show="showHelp"
      @close="showHelp = false"
    />
  </div>
</template>

<script>
import BulkLinkAdder from './components/BulkLinkAdder.vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import KeyboardShortcuts from './components/KeyboardShortcuts.vue'

export default {
  name: 'App',
  components: {
    BulkLinkAdder,
    FontAwesomeIcon,
    KeyboardShortcuts
  },
  data() {
    return {
      files: [],
      subjects: [],
      currentFile: '',
      searchQuery: '',
      filteredSubjects: [],
      showBulkDialog: false,
      currentSubject: '',
      focusedLinkIndex: -1,
      currentSubjectIndex: -1,
      showHelp: false,
      keyboardShortcuts: {
        'x': 'Toggle selection',
        'Enter': 'Open link in new tab',
        'Delete': 'Delete selected links',
        'Shift + #': 'Delete selected links',
        'ArrowUp/k': 'Previous link',
        'ArrowDown/j': 'Next link',
        'c': 'Copy selected links',
      }
    }
  },
  computed: {
    hasLinks() {
      return this.subjects.some(subject => subject.links && subject.links.length > 0)
    },
    selectedCount() {
      return this.subjects.reduce((count, subject) => 
        count + (subject.links ? subject.links.filter(link => link.selected).length : 0), 0)
    },
    totalLinks() {
      return this.subjects.reduce((count, subject) => 
        count + (subject.links ? subject.links.length : 0), 0)
    },
    allSelected() {
      return this.hasLinks && this.selectedCount === this.totalLinks
    }
  },
  mounted() {
    this.loadFiles()
    
    // Add global keyboard event listener
    document.addEventListener('keydown', this.handleGlobalKeydown)

    // Add keyboard shortcuts for file and search focus
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey && e.key === 'o') {
        e.preventDefault()
        document.getElementById('file-list').focus()
      } else if (e.ctrlKey && e.key === 'f') {
        e.preventDefault()
        document.getElementById('search-input').focus()
      }
    })

    // Add global shortcut for help panel
    document.addEventListener('keydown', (e) => {
      if (e.key === '?' && !e.ctrlKey && !e.altKey) {
        e.preventDefault()
        this.showHelp = !this.showHelp
      }
    })

    // Focus first link after loading
    this.$nextTick(() => {
      this.focusFirstLink()
    })
  },
  beforeUnmount() {
    // Clean up event listeners
    document.removeEventListener('keydown', this.handleGlobalKeydown)
  },
  methods: {
    async loadFiles() {
      try {
        const response = await fetch('/api/files')
        const data = await response.json()
        this.files = data
        
        // Check URL parameters
        const urlParams = new URLSearchParams(window.location.search)
        const fileParam = urlParams.get('file')
        if (fileParam && this.files.includes(fileParam)) {
          this.currentFile = fileParam
          await this.loadLinks()
        }
      } catch (error) {
        console.error('Error loading files:', error)
      }
    },
    async loadLinks() {
      if (!this.currentFile) return
      
      try {
        const response = await fetch(`/api/file/${this.currentFile}`)
        const data = await response.json()
        this.subjects = data.data || []
        // Initialize selected property for each link and ensure links array exists
        this.subjects.forEach(subject => {
          if (!subject.links) {
            subject.links = []
          }
          subject.links.forEach(link => {
            link.selected = false
          })
        })
        this.searchLinks(this.searchQuery)
        this.updateUrl(this.currentFile, this.searchQuery)
        
        // Focus first link after loading
        this.$nextTick(() => {
          this.focusFirstLink()
        })
      } catch (error) {
        console.error('Error loading links:', error)
      }
    },
    updateUrl(filename, searchQuery = '') {
      const newUrl = new URL(window.location)
      newUrl.searchParams.set('file', filename)
      if (searchQuery) {
        newUrl.searchParams.set('search', searchQuery)
      } else {
        newUrl.searchParams.delete('search')
      }
      window.history.pushState({}, '', newUrl)
    },
    calculateRelevance(link, keywords) {
      const titleWords = link.title.toLowerCase().split(/\s+/)
      const urlWords = link.url.toLowerCase().split(/\s+/)
      
      let relevance = 0
      let matchedKeywords = 0

      for (const keyword of keywords) {
        let keywordMatched = false

        for (const word of titleWords) {
          if (word === keyword) {
            relevance += 2
            keywordMatched = true
            break
          } else if (word.includes(keyword)) {
            relevance += 1
            keywordMatched = true
            break
          }
        }

        if (!keywordMatched) {
          for (const word of urlWords) {
            if (word === keyword) {
              relevance += 1
              keywordMatched = true
              break
            } else if (word.includes(keyword)) {
              relevance += 0.5
              keywordMatched = true
              break
            }
          }
        }

        if (keywordMatched) matchedKeywords++
      }

      return matchedKeywords === keywords.length ? relevance : 0
    },
    searchLinks(searchQuery) {
      if (!searchQuery) {
        this.filteredSubjects = this.subjects
        return
      }

      const keywords = searchQuery.toLowerCase().split(/\s+/).filter(keyword => keyword.length > 0)

      if (keywords.length === 0) {
        this.filteredSubjects = this.subjects
        return
      }

      this.filteredSubjects = this.subjects
        .map(subject => {
          const relevantLinks = subject.links
            .map(link => {
              const relevance = this.calculateRelevance(link, keywords)
              return relevance > 0 ? { ...link, relevance } : null
            })
            .filter(Boolean)
            .sort((a, b) => b.relevance - a.relevance)

          return relevantLinks.length > 0
            ? { ...subject, links: relevantLinks }
            : null
        })
        .filter(Boolean)
        .sort((a, b) => b.links[0].relevance - a.links[0].relevance)
    },
    async deleteSelected() {
      const selectedLinks = this.subjects
        .flatMap(subject => 
          subject.links
            .filter(link => link.selected)
            .map(link => ({
              title: link.title,
              url: link.url
            }))
        )

      if (selectedLinks.length === 0) {
        alert('Please select at least one link to delete.')
        return
      }

      if (confirm(`Are you sure you want to delete ${selectedLinks.length} selected link(s)?`)) {
        try {
          const response = await fetch('/api/delete_links', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify({
              filename: this.currentFile,
              links: selectedLinks
            }),
          })

          if (response.ok) {
            await this.loadLinks()
          } else {
            throw new Error('Failed to delete links')
          }
        } catch (error) {
          console.error('Error deleting links:', error)
          alert('Failed to delete links. Please try again.')
        }
      }
    },
    toggleAllFiltered(event) {
      const checked = event.target.checked
      this.subjects.forEach(subject => {
        subject.links.forEach(link => {
          link.selected = checked
        })
      })
    },
    async copyFilteredLinks() {
      const selectedLinks = this.subjects
        .flatMap(subject =>
          subject.links
            .filter(link => link.selected)
            .map(link => `- [${link.title}](${link.url})`)
        )

      if (selectedLinks.length === 0) {
        alert('Please select at least one link to copy.')
        return
      }

      try {
        await navigator.clipboard.writeText(selectedLinks.join('\n'))
        alert('Selected links copied to clipboard!')
      } catch (err) {
        console.error('Failed to copy links: ', err)
        alert('Failed to copy links. Please try again.')
      }
    },
    onCheckboxChange() {
      // This method can be extended if needed
    },
    onSearchInput: debounce(function() {
      this.searchLinks(this.searchQuery)
      this.updateUrl(this.currentFile, this.searchQuery)
    }, 300),
    async handleLinksAdded() {
      this.showBulkDialog = false
      await this.loadLinks()
    },
    openBulkAddDialog(subject) {
      this.currentSubject = subject
      this.showBulkDialog = true
    },
    handleGlobalKeydown(e) {
      // Skip if we're in an input field
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return

      if (e.shiftKey) {
        switch (e.key) {
          case '#':
            e.preventDefault()
            this.deleteSelected()
            break
          case 'C':
            e.preventDefault()
            this.copyFilteredLinks()
            break
          case 'N':
            e.preventDefault()
            this.openBulkAddDialog('')
            break
        }
        return
      }

      switch (e.key) {
        case 'x':
          e.preventDefault()
          // Get the currently focused link element
          const focusedElement = document.activeElement
          if (focusedElement && focusedElement.classList.contains('link-item')) {
            // Find the link data that corresponds to this element
            const url = focusedElement.getAttribute('data-url')
            const link = this.subjects
              .flatMap(subject => subject.links)
              .find(link => link.url === url)
            
            if (link) {
              link.selected = !link.selected
              this.onCheckboxChange(link)
            }
          }
          break
        case 'Enter':
          e.preventDefault()
          if (this.selectedCount > 0) {
            this.openSelectedLinks()
          } else {
            this.openLink(this.getVisibleLinks()[this.focusedLinkIndex])
          }
          break
        case 'ArrowDown':
        case 'j':
          e.preventDefault()
          this.focusNextLink(this.getAllLinks().indexOf(e.target))
          break
        case 'ArrowUp':
        case 'k':
          e.preventDefault()
          this.focusPreviousLink(this.getAllLinks().indexOf(e.target))
          break
        case 'Delete':
          e.preventDefault()
          this.deleteSelected()
          break
      }
    },
    openLink(link) {
      window.open(link.url, '_blank')
    },
    openSelectedLinks() {
      const selectedLinks = this.subjects
        .flatMap(subject => 
          subject.links.filter(link => link.selected)
        )
      
      selectedLinks.forEach(link => {
        window.open(link.url, '_blank')
      })
    },
    focusNextLink(currentIndex) {
      const allLinks = this.getAllLinks()
      if (allLinks.length === 0) return

      const nextIndex = (currentIndex + 1) % allLinks.length
      allLinks[nextIndex]?.focus()
    },
    focusPreviousLink(currentIndex) {
      const allLinks = this.getAllLinks()
      if (allLinks.length === 0) return

      const prevIndex = currentIndex - 1 < 0 ? allLinks.length - 1 : currentIndex - 1
      allLinks[prevIndex]?.focus()
    },
    getAllLinks() {
      return Array.from(document.querySelectorAll('.link-item'))
        .filter(link => {
          const style = window.getComputedStyle(link)
          return style.display !== 'none' && style.visibility !== 'hidden'
        })
    },
    focusFirstLink() {
      const firstLink = this.getAllLinks()[0]
      if (firstLink) {
        firstLink.focus()
      }
    },
    getVisibleLinks() {
      return this.subjects
        .flatMap(subject => subject.links)
        .filter(link => {
          const style = window.getComputedStyle(document.querySelector(`.link-item[data-url="${link.url}"]`))
          return style.display !== 'none' && style.visibility !== 'hidden'
        })
    },
  },
  watch: {
    currentFile() {
      this.loadLinks()
    }
  }
}

function debounce(func, delay) {
  let timeoutId
  return function(...args) {
    clearTimeout(timeoutId)
    timeoutId = setTimeout(() => func.apply(this, args), delay)
  }
}
</script>

<style>
:root {
  --primary-color: #3498db;
  --secondary-color: #2c3e50;
  --background-color: #f5f5f5;
  --card-background: #ffffff;
  --text-color: #333333;
  --border-color: #e0e0e0;
  --danger-color: #e74c3c;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background-color: var(--background-color);
  color: var(--text-color);
  font-size: 14px;
}

.app {
  max-width: 1000px;
  margin: 0 auto;
  padding: 20px;
}

.controls {
  padding: 5px 0;
  display: flex;
  gap: 10px;
  margin: 10px 0;
}

#file-list, #search-input {
  flex: 1;
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 3px;
  font-size: 14px;
  background-color: var(--card-background);
  color: var(--text-color);
  transition: border-color 0.3s, box-shadow 0.3s;
}

#file-list:focus, #search-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

#preview-wrapper {
  background-color: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  padding: 15px;
  box-shadow: 0 2px 5px rgba(0,0,0,0.1);
  display: flex;
  gap: 20px;
  overflow: hidden;
  width: calc(100% - 32px);
}

.main-content {
  flex: 1;
  min-width: 0;
  overflow-x: hidden;
}

.multi-select-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.control-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.control-buttons button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 28px;
  width: 28px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
  color: #666;
  margin: 0;
}

.control-buttons input[type="checkbox"] {
  margin: 0;
  width: 14px;
  height: 14px;
  cursor: pointer;
}

.control-buttons button:hover {
  background-color: #f0f0f0;
}

#delete-selected-btn:hover {
  color: var(--danger-color);
}

#copy-filtered-btn:hover {
  color: var(--secondary-color);
}

#bulk-add-btn:hover {
  color: var(--primary-color);
}

#add-to-subject-btn:hover {
  color: var(--secondary-color);
}

.subject-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  min-height: 28px;
}

.subject {
  font-weight: bold;
  font-size: 1.1em;
  line-height: 28px;
  padding: 0;
  margin: 0;
}

.add-to-subject-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
  color: #666;
  height: 28px;
  margin: 0;
}

.subject {
  font-size: 1.2em;
  font-weight: bold;
  margin: 20px 0 10px;
  color: var(--secondary-color);
}

.links-list {
  margin-bottom: 20px;
}

.link-item {
  display: flex;
  align-items: center;
  padding: 5px 0;
  border-bottom: 1px solid var(--border-color);
  border-left: 3px solid transparent;
  overflow: hidden;
  width: 100%;
  outline: none;
  transition: all 0.2s ease;
}

.link-item:last-child {
  border-bottom: none;
}

.link-item p {
  margin: 0;
  flex-grow: 1;
  font-size: 1em;
  line-height: 1.4;
  padding-left: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-item a {
  color: #01579b;
  text-decoration: none;
}

.link-item a:hover {
  text-decoration: underline;
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  width: 25px;
}

#delete-selected-btn, #copy-filtered-btn, #bulk-add-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  border-radius: 3px;
  transition: background-color 0.3s;
}

#delete-selected-btn:hover {
  background-color: #FFC5C5;
}

#copy-filtered-btn:hover {
  background-color: #D3D3D3;
}

#bulk-add-btn:hover {
  background-color: #C5FFC5;
}

.add-to-subject-btn:hover {
  background-color: #D3D3D3;
}

.link-item.selected {
  background-color: rgba(52, 152, 219, 0.1);
}

.link-item.selected:hover {
  background-color: rgba(52, 152, 219, 0.15);
}

.link-item.selected:focus {
  background-color: rgba(52, 152, 219, 0.2);
}

.link-item.selected a {
  text-decoration: none;
}

.no-links {
  text-align: center;
  padding: 20px;
  color: #666;
  font-style: italic;
}

.icon-button {
  padding: 8px;
  background: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 35px;
  height: 35px;
  color: var(--text-color);
  transition: all 0.2s ease;
}

.icon-button:hover {
  background: var(--background-color);
  color: var(--primary-color);
}

.icon-button i {
  font-size: 14px;
}

/* Add styles for keyboard focus */
.link-item:focus {
  background-color: rgba(52, 152, 219, 0.08);
  border-left: 3px solid rgba(52, 152, 219, 0.5);
}

/* Optional: Add a subtle left border to indicate focus */
.link-item {
  transition: background-color 0.2s ease, border-left-color 0.2s ease;
}

/* Make sure the hover state is distinct but complementary */
.link-item:hover {
  background-color: rgba(0, 0, 0, 0.02);
}

/* Ensure the focus state is visible even when link is selected */
.link-item.selected:focus {
  border-left: 3px solid rgba(52, 152, 219, 0.5);
}

/* Optional: Add a help tooltip for keyboard shortcuts */
.keyboard-shortcuts-help {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 15px rgba(0,0,0,0.2);
  font-size: 0.9em;
  color: #666;
  max-width: 400px;
  z-index: 1000;
}

.shortcuts-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.shortcuts-header h3 {
  margin: 0;
  color: var(--secondary-color);
}

.shortcuts-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.shortcut-item kbd {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 3px;
  border: 1px solid #ddd;
  box-shadow: 0 1px 1px rgba(0,0,0,0.2);
  font-family: monospace;
  font-size: 0.9em;
  min-width: 20px;
  text-align: center;
}

.shortcut-item span {
  flex: 1;
}

.keyboard-shortcuts-help .close-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  padding: 5px;
  color: #666;
  border-radius: 50%;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.keyboard-shortcuts-help .close-btn:hover {
  background: #f0f0f0;
  color: #333;
}

.help-button {
  position: fixed;
  bottom: 20px;
  right: 20px;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #3498db;
  color: white;
  border: none;
  font-size: 20px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0,0,0,0.15);
  transition: all 0.2s;
}

.help-button:hover {
  background: #2980b9;
  transform: scale(1.05);
}

.help-button:active {
  transform: scale(0.95);
}
</style>
