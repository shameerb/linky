<!-- App.vue -->
<template>
  <div class="app">
    <div class="app-container">
      <div class="controls">
        <select v-model="currentFile" id="file-list">
          <option value="">Select a file</option>
          <option v-for="file in files" :key="file" :value="file">
            {{ file }}
          </option>
        </select>
        
        <div class="search-container">
          <input
            type="text"
            v-model="searchQuery"
            @input="onSearchInput"
            @keydown="handleSearchKeydown"
            placeholder="Search links... (use subject:<name> to filter by subject)"
            id="search-input"
            ref="searchInput"
          />
          
          <!-- Subject suggestions dropdown -->
          <div v-if="showSubjectSuggestions && subjectSuggestions.length > 0" class="subject-suggestions">
            <div 
              v-for="(subject, index) in subjectSuggestions" 
              :key="subject"
              class="subject-suggestion"
              :class="{ 'selected': index === selectedSuggestionIndex }"
              @click="selectSubjectSuggestion(subject)"
              @mouseover="selectedSuggestionIndex = index"
            >
              {{ subject }}
            </div>
          </div>
        </div>
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
                :checked="allFilteredSelected"
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
              <span v-if="selectedFilteredCount > 0" class="selected-count">
                Selected: {{ selectedFilteredCount }}
              </span>
              Total: <span id="total-links">{{ filteredLinksCount }}</span>
            </span>
          </div>

          <div 
            v-for="subject in filteredSubjects" 
            :key="subject.subject" 
            class="subject-group"
            :class="{ 'collapsed': isSubjectCollapsed(subject.subject) }"
          >
            <div class="subject-header" @click="toggleSubjectCollapse(subject.subject)">
              <div class="collapse-icon">
                {{ isSubjectCollapsed(subject.subject) ? '▶' : '▼' }}
              </div>
              <div class="subject">{{ subject.subject }}</div>
              <div class="link-count">({{ subject.links.length }})</div>
            </div>
            <div class="links-list" v-show="!isSubjectCollapsed(subject.subject)">
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

    <!-- Notification -->
    <div v-if="notification.show" class="notification" :class="notification.type">
      {{ notification.message }}
    </div>
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
      currentSubjectFilter: '', // Subject filter
      showSubjectSuggestions: false,
      subjectSearchQuery: '',
      keyboardShortcuts: {
        'x': 'Toggle selection',
        'Enter': 'Open link in new tab',
        'Delete': 'Delete selected links',
        'Shift + #': 'Delete selected links',
        'ArrowUp/k': 'Previous link',
        'ArrowDown/j': 'Next link',
      },
      selectedSuggestionIndex: -1, // Track the currently selected suggestion
      notification: {
        show: false,
        message: '',
        type: 'info',
        timeout: null
      },
      collapsedSubjects: new Set(), // Track which subjects are collapsed
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
    },
    availableSubjects() {
      return this.subjects.map(s => s.subject).filter(Boolean)
    },
    // Extract the search part without the subject filter
    searchQueryWithoutSubject() {
      // Remove any existing subject: prefix
      return this.searchQuery.replace(/subject:[^\s]+\s*/, '').trim()
    },
    // Get subject suggestions based on current input
    subjectSuggestions() {
      if (!this.showSubjectSuggestions) return []
      
      // Extract the partial subject name after "subject:"
      const match = this.searchQuery.match(/subject:([^&\s]*)/)
      if (!match) return []
      
      const partialSubject = match[1].toLowerCase()
      if (!partialSubject) return this.availableSubjects
      
      // Filter subjects that match the partial input
      return this.availableSubjects.filter(subject => 
        subject.toLowerCase().includes(partialSubject)
      )
    },
    // Update to count only filtered links
    filteredLinksCount() {
      return this.filteredSubjects.reduce((count, subject) => 
        count + (subject.links ? subject.links.length : 0), 0)
    },
    
    // Update to check if all filtered links are selected
    allFilteredSelected() {
      if (!this.hasFilteredLinks) return false
      
      const selectedFilteredCount = this.filteredSubjects.reduce((count, subject) => 
        count + (subject.links ? subject.links.filter(link => link.selected).length : 0), 0)
      
      return selectedFilteredCount === this.filteredLinksCount
    },
    
    // Add a computed property to check if there are any filtered links
    hasFilteredLinks() {
      return this.filteredSubjects.some(subject => subject.links && subject.links.length > 0)
    },
    selectedFilteredCount() {
      return this.filteredSubjects.reduce((count, subject) => 
        count + (subject.links ? subject.links.filter(link => link.selected).length : 0), 0)
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

    // Close subject suggestions when clicking outside
    document.addEventListener('click', this.closeSubjectSuggestions)

    // Focus first link after loading
    this.$nextTick(() => {
      this.focusFirstLink()
    })
  },
  beforeUnmount() {
    // Clean up event listeners
    document.removeEventListener('keydown', this.handleGlobalKeydown)
    document.removeEventListener('click', this.closeSubjectSuggestions)
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
        
        // Set default file to now.md if it exists in the files list
        if (fileParam && this.files.includes(fileParam)) {
          this.currentFile = fileParam
        } else if (this.files.includes('now.md')) {
          this.currentFile = 'now.md'
        } else if (this.files.length > 0) {
          this.currentFile = this.files[0]
        }
        
        if (this.currentFile) {
          await this.loadLinks()
        }
        
        // Check for subject filter and search in URL
        const subjectParam = urlParams.get('subject')
        const searchParam = urlParams.get('search')
        
        // Build search query from URL parameters
        let newSearchQuery = ''
        
        if (subjectParam) {
          this.currentSubjectFilter = subjectParam
          newSearchQuery = `subject:${subjectParam}`
        }
        
        if (searchParam) {
          if (newSearchQuery) {
            newSearchQuery += ` & ${searchParam}`
          } else {
            newSearchQuery = searchParam
          }
        }
        
        if (newSearchQuery) {
          this.searchQuery = newSearchQuery
          this.applyFilters()
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
        
        this.applyFilters()
        
        // Focus first link after loading
        this.$nextTick(() => {
          this.focusFirstLink()
        })
      } catch (error) {
        console.error('Error loading links:', error)
      }
    },
    updateUrl() {
      const newUrl = new URL(window.location)
      newUrl.searchParams.set('file', this.currentFile)
      
      // Update search parameter
      if (this.searchQueryWithoutSubject) {
        newUrl.searchParams.set('search', this.searchQueryWithoutSubject)
      } else {
        newUrl.searchParams.delete('search')
      }
      
      // Update subject parameter
      if (this.currentSubjectFilter) {
        newUrl.searchParams.set('subject', this.currentSubjectFilter)
      } else {
        newUrl.searchParams.delete('subject')
      }
      
      window.history.pushState({}, '', newUrl)
    },
    applyFilters() {
      // Parse subject filter from search query
      const subjectMatch = this.searchQuery.match(/subject:([^\s]+)/)
      if (subjectMatch) {
        this.currentSubjectFilter = subjectMatch[1]
      } else {
        this.currentSubjectFilter = ''
      }
      
      // Get search terms (everything after the subject: part if it exists)
      let searchTerms = ''
      if (subjectMatch) {
        // Extract everything after the subject: part
        const afterSubject = this.searchQuery.substring(
          this.searchQuery.indexOf(subjectMatch[0]) + subjectMatch[0].length
        ).trim()
        searchTerms = afterSubject
      } else {
        searchTerms = this.searchQuery.trim()
      }
      
      // Apply filters
      this.filterLinks(this.currentSubjectFilter, searchTerms)
      
      // Update URL
      this.updateUrl()
    },
    filterLinks(subjectFilter, searchTerms) {
      // First filter by subject if needed
      let filteredBySubject = this.subjects
      if (subjectFilter) {
        filteredBySubject = this.subjects.filter(subject => 
          subject.subject === subjectFilter
        )
      }
      
      // If no search terms after filtering by subject, return all subjects
      if (!searchTerms) {
        this.filteredSubjects = filteredBySubject
        return
      }
      
      // Then apply search filter
      const keywords = searchTerms.toLowerCase().split(/\s+/).filter(keyword => keyword.length > 0)
      
      if (keywords.length === 0) {
        this.filteredSubjects = filteredBySubject
        return
      }
      
      this.filteredSubjects = filteredBySubject
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
      
      // Boost relevance if all keywords matched
      if (matchedKeywords === keywords.length && keywords.length > 0) {
        relevance *= 1.5
      }
      
      return relevance
    },
    onSearchInput: debounce(function() {
      // Check if we're typing in the subject: part
      if (this.searchQuery.match(/subject:[^&\s]*$/)) {
        this.showSubjectSuggestions = true
        this.selectedSuggestionIndex = -1 // Reset selection index when input changes
      } else {
        this.showSubjectSuggestions = false
      }
      
      this.applyFilters()
    }, 300),
    handleSearchKeydown(e) {
      // Handle navigation in subject suggestions
      if (this.showSubjectSuggestions && this.subjectSuggestions.length > 0) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          this.selectedSuggestionIndex = Math.min(
            this.selectedSuggestionIndex + 1, 
            this.subjectSuggestions.length - 1
          )
          if (this.selectedSuggestionIndex === -1) {
            this.selectedSuggestionIndex = 0
          }
        } else if (e.key === 'ArrowUp') {
          e.preventDefault()
          this.selectedSuggestionIndex = Math.max(this.selectedSuggestionIndex - 1, 0)
        } else if (e.key === 'Enter' || e.key === 'Tab') {
          e.preventDefault()
          if (this.subjectSuggestions.length > 0) {
            // If a suggestion is highlighted, select it
            if (this.selectedSuggestionIndex >= 0) {
              this.selectSubjectSuggestion(this.subjectSuggestions[this.selectedSuggestionIndex])
            } else {
              // Otherwise select the first suggestion
              this.selectSubjectSuggestion(this.subjectSuggestions[0])
            }
          }
        } else if (e.key === 'Escape') {
          e.preventDefault()
          this.showSubjectSuggestions = false
          this.selectedSuggestionIndex = -1
        }
      }
    },
    selectSubjectSuggestion(subject) {
      // Replace the partial subject with the selected one
      const beforeSubject = this.searchQuery.split('subject:')[0]
      let afterSubject = ''
      
      // Check if there's content after the subject
      const afterMatch = this.searchQuery.match(/subject:[^\s]*(.*?)$/)
      if (afterMatch && afterMatch[1]) {
        afterSubject = afterMatch[1]
      }
      
      this.searchQuery = `${beforeSubject}subject:${subject}${afterSubject}`
      this.showSubjectSuggestions = false
      this.selectedSuggestionIndex = -1 // Reset selection index
      this.applyFilters()
      
      // Keep focus on the search input
      this.$nextTick(() => {
        this.$refs.searchInput.focus()
      })
    },
    closeSubjectSuggestions(event) {
      // Don't close if clicking inside the search input
      if (this.$refs.searchInput && this.$refs.searchInput.contains(event.target)) {
        return
      }
      
      this.showSubjectSuggestions = false
      this.selectedSuggestionIndex = -1 // Reset selection index
    },
    async deleteSelected() {
      // Get all visible links from filtered subjects
      const visibleLinks = this.getVisibleLinks()
      
      // Filter to only selected links
      const selectedLinks = visibleLinks.filter(link => link.selected)
      
      if (selectedLinks.length === 0) {
        this.showNotification('No links selected', 'error')
        return
      }
      
      if (!confirm(`Delete ${selectedLinks.length} selected links?`)) {
        return
      }
      
      try {
        const response = await fetch('/api/delete_links', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            filename: this.currentFile,
            links: selectedLinks.map(link => ({
              id: link.id,
              title: link.title,
              url: link.url
            }))
          })
        })
        
        if (response.ok) {
          this.showNotification(`Deleted ${selectedLinks.length} links`, 'success')
          // Reload links after deletion
          await this.loadLinks()
        } else {
          this.showNotification('Failed to delete links', 'error')
          console.error('Failed to delete links')
        }
      } catch (error) {
        this.showNotification('Error deleting links', 'error')
        console.error('Error deleting links:', error)
      }
    },
    toggleAllFiltered() {
      const shouldSelect = !this.allFilteredSelected
      
      // Directly update the filtered subjects
      this.filteredSubjects.forEach(subject => {
        subject.links.forEach(link => {
          // Find the original link in the subjects array and update it
          for (const originalSubject of this.subjects) {
            const originalLink = originalSubject.links.find(l => l.url === link.url)
            if (originalLink) {
              originalLink.selected = shouldSelect
              // Also update the filtered link to keep UI in sync
              link.selected = shouldSelect
            }
          }
        })
      })
    },
    copyFilteredLinks() {
      // Get all visible links from filtered subjects
      const visibleLinks = this.getVisibleLinks()
      
      // Filter to only selected links
      const selectedLinks = visibleLinks.filter(link => link.selected)
      
      if (selectedLinks.length === 0) {
        this.showNotification('No links selected', 'error')
        return
      }
      
      const text = selectedLinks
        .map(link => `[${link.title}](${link.url})`)
        .join('\n')
      
      navigator.clipboard.writeText(text)
        .then(() => {
          this.showNotification(`Copied ${selectedLinks.length} links to clipboard`, 'success')
        })
        .catch(err => {
          console.error('Failed to copy links: ', err)
          this.showNotification('Failed to copy links', 'error')
        })
    },
    onCheckboxChange(changedLink) {
      // Create a map of URLs for faster lookup
      const visibleLinks = this.getVisibleLinks()
      const visibleUrlMap = new Set(visibleLinks.map(link => link.url))
      
      // Only update if the link is visible in the filtered view
      if (visibleUrlMap.has(changedLink.url)) {
        // Update all instances of this link in both original and filtered data
        this.subjects.forEach(subject => {
          subject.links.forEach(link => {
            if (link.url === changedLink.url) {
              link.selected = changedLink.selected
            }
          })
        })
        
        this.filteredSubjects.forEach(subject => {
          subject.links.forEach(link => {
            if (link.url === changedLink.url) {
              link.selected = changedLink.selected
            }
          })
        })
      }
    },
    openBulkAddDialog(subject) {
      this.currentSubject = subject
      this.showBulkDialog = true
    },
    async handleLinksAdded() {
      this.showBulkDialog = false
      await this.loadLinks()
    },
    openLink(link) {
      window.open(link.url, '_blank')
    },
    focusFirstLink() {
      const firstLink = this.getAllLinks()[0]
      if (firstLink) {
        firstLink.focus()
      }
    },
    getAllLinks() {
      return Array.from(document.querySelectorAll('.link-item'))
        .filter(link => {
          const style = window.getComputedStyle(link)
          return style.display !== 'none' && style.visibility !== 'hidden'
        })
    },
    handleGlobalKeydown(e) {
      // Skip if we're in an input field
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return
      
      // Skip if Command/Ctrl key is pressed to allow browser shortcuts
      if (e.metaKey || e.ctrlKey) return

      // Get the currently focused link element once
      const focusedElement = document.activeElement

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
          case '|':  // Add this case for opening links
            e.preventDefault()
            if (this.selectedCount > 0) {
              this.openSelectedLinks()
            }
            break
        }
        return
      }

      switch (e.key) {
        case 'x':
          e.preventDefault()
          if (focusedElement && focusedElement.classList.contains('link-item')) {
            // Find the link data that corresponds to this element
            const url = focusedElement.getAttribute('data-url')
            
            // Search through filtered subjects first for better performance
            let found = false
            for (const subject of this.filteredSubjects) {
              for (const link of subject.links) {
                if (link.url === url) {
                  link.selected = !link.selected
                  this.onCheckboxChange(link)
                  found = true
                  break
                }
              }
              if (found) break
            }
            
            // If not found in filtered subjects, search all subjects
            if (!found) {
              for (const subject of this.subjects) {
                for (const link of subject.links) {
                  if (link.url === url) {
                    link.selected = !link.selected
                    this.onCheckboxChange(link)
                    break
                  }
                }
              }
            }
          }
          break
        case 'Enter':
          e.preventDefault()
          if (focusedElement && focusedElement.classList.contains('link-item')) {
            const url = focusedElement.getAttribute('data-url')
            const link = this.getVisibleLinks().find(l => l.url === url)
            if (link) {
              this.openLink(link)
            }
          }
          break
        case 'ArrowDown':
        case 'j':
          e.preventDefault()
          this.focusNextLink(this.getAllLinks().indexOf(document.activeElement))
          break
        case 'ArrowUp':
        case 'k':
          e.preventDefault()
          this.focusPreviousLink(this.getAllLinks().indexOf(document.activeElement))
          break
        case 'Delete':
          e.preventDefault()
          this.deleteSelected()
          break
      }
    },
    openSelectedLinks() {
      // Get all visible links from filtered subjects
      const visibleLinks = this.getVisibleLinks()
      
      // Filter to only selected links
      const selectedLinks = visibleLinks.filter(link => link.selected)
      
      if (selectedLinks.length === 0) {
        this.showNotification('No links selected', 'error')
        return
      }
      
      console.log('Selected links:', selectedLinks)
      
      try {
        // Open links in batches of 5
        const BATCH_SIZE = 5
        const batches = []
        
        // Split links into batches
        for (let i = 0; i < selectedLinks.length; i += BATCH_SIZE) {
          batches.push(selectedLinks.slice(i, i + BATCH_SIZE))
        }
        
        // Open each batch with a small delay between batches
        batches.forEach((batch, index) => {
          setTimeout(() => {
            batch.forEach(link => window.open(link.url, '_blank'))
          }, index * 100) // 100ms delay between batches
        })
        
        this.showNotification(`Opening ${selectedLinks.length} links in batches`, 'success')
      } catch (error) {
        console.error('Error opening links:', error)
        this.showNotification('Error opening links', 'error')
      }
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
    getVisibleLinks() {
      // First get all links from filtered subjects
      return this.filteredSubjects.flatMap(subject => subject.links || [])
    },
    showNotification(message, type = 'info', duration = 3000) {
      // Clear any existing timeout
      if (this.notification.timeout) {
        clearTimeout(this.notification.timeout)
      }
      
      // Show new notification
      this.notification.message = message
      this.notification.type = type
      this.notification.show = true
      
      // Auto-hide after duration
      this.notification.timeout = setTimeout(() => {
        this.notification.show = false
      }, duration)
    },
    // Toggle collapse state for a subject
    toggleSubjectCollapse(subject) {
      if (this.collapsedSubjects.has(subject)) {
        this.collapsedSubjects.delete(subject)
      } else {
        this.collapsedSubjects.add(subject)
      }
    },
    
    // Check if a subject is collapsed
    isSubjectCollapsed(subject) {
      return this.collapsedSubjects.has(subject)
    },
    
    // Update getAllLinks to only include links from expanded subjects
    getAllLinks() {
      return Array.from(document.querySelectorAll('.link-item'))
        .filter(link => {
          const style = window.getComputedStyle(link)
          return style.display !== 'none' && style.visibility !== 'hidden'
        })
    },
  },
  watch: {
    currentFile() {
      this.loadLinks()
      // Reset filters when changing files
      this.searchQuery = ''
      this.currentSubjectFilter = ''
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
  padding: 20px;
  width: 100%;
  box-sizing: border-box;
}

.app-container {
  max-width: 1000px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
}

.controls {
  padding: 5px 0;
  display: flex;
  gap: 10px;
  margin: 10px 0;
  width: 100%;
  box-sizing: border-box;
}

#file-list {
  flex: 0.3;
  min-width: 120px;
  max-width: 30%;
}

.search-container {
  position: relative;
  flex: 0.7;
  width: 70%;
  box-sizing: border-box;
}

#file-list {
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 3px;
  font-size: 14px;
  background-color: var(--card-background);
  color: var(--text-color);
  transition: border-color 0.3s, box-shadow 0.3s;
}

#search-input {
  width: 100%;
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 3px;
  font-size: 14px;
  background-color: var(--card-background);
  color: var(--text-color);
  transition: border-color 0.3s, box-shadow 0.3s;
  box-sizing: border-box;
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
  width: 100%;
  box-sizing: border-box;
  min-height: 400px; /* Set minimum height */
}

.main-content {
  flex: 1;
  min-width: 0;
  overflow-x: hidden;
  width: 100%;
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
  margin-bottom: 5px;
  min-height: 24px;
  cursor: pointer;
  padding: 3px 5px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.subject-header:hover {
  background-color: rgba(0, 0, 0, 0.05);
}

.collapse-icon {
  font-size: 10px;
  color: var(--secondary-color);
  width: 14px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.link-count {
  color: #666;
  font-size: 0.8em;
  margin-left: 5px;
}

.subject {
  font-size: 1em;
  font-weight: bold;
  margin: 0;
  color: var(--secondary-color);
}

.links-list {
  margin-bottom: 10px;
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

.subject-suggestions {
  position: absolute;
  top: 100%;
  left: 0;
  width: 100%;
  max-height: 200px;
  overflow-y: auto;
  background-color: var(--card-background);
  border: 1px solid var(--border-color);
  border-radius: 0 0 4px 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  z-index: 10;
}

.subject-suggestion {
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.subject-suggestion:hover,
.subject-suggestion.selected {
  background-color: rgba(52, 152, 219, 0.1);
}

/* Add styles for the selected count */
.selected-count {
  margin-right: 15px;
  font-weight: 500;
  color: var(--primary-color);
}

/* Add styles for notifications */
.notification {
  position: fixed;
  bottom: 80px;
  right: 20px;
  padding: 12px 20px;
  border-radius: 4px;
  color: white;
  font-size: 14px;
  z-index: 1000;
  box-shadow: 0 2px 10px rgba(0,0,0,0.2);
  animation: fadeIn 0.3s, fadeOut 0.3s 2.7s;
  max-width: 300px;
}

.notification.info {
  background-color: var(--primary-color);
}

.notification.success {
  background-color: #2ecc71;
}

.notification.error {
  background-color: var(--danger-color);
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes fadeOut {
  from { opacity: 1; transform: translateY(0); }
  to { opacity: 0; transform: translateY(-10px); }
}

/* Add styles for collapsible subjects */
.subject-group {
  margin-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 5px;
}

.subject-group:last-child {
  border-bottom: none;
}

.subject-group.collapsed {
  margin-bottom: 3px;
  padding-bottom: 3px;
}
</style>
