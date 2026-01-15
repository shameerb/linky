<!-- App.vue -->
<template>
  <div class="app">
    <!-- Show login if not authenticated -->
    <Login v-if="!isAuthenticated" @login="handleLogin" />

    <!-- Show main app if authenticated -->
    <div v-else class="app-container">
      <!-- Header with Linky title and logout -->
      <div class="app-header" :class="{ 'header-hidden': isHeaderHidden }">
        <h1 class="app-title">Linky</h1>
        <button class="logout-btn" @click="logout" title="Logout">
          Logout
        </button>
      </div>

      <div class="multi-select-controls" :class="{ 'controls-hidden': isHeaderHidden }">
        <div class="action-controls">
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
              :disabled="!currentSubjectId"
            >
              <font-awesome-icon icon="plus" />
            </button>
          </div>

          <div class="search-controls">
            <select v-model="currentSubjectId" id="subject-list" @change="loadSubjectData">
              <option value="">All Subjects</option>
              <option v-for="subject in subjects" :key="subject.id" :value="subject.id">
                {{ subject.name }}
              </option>
            </select>

            <div class="search-container">
              <input
                type="text"
                v-model="searchQuery"
                @input="onSearchInput"
                @keydown="handleSearchKeydown"
                placeholder="Search links... (Use topic:<name>, OR, AND, & operators)"
                id="search-input"
                ref="searchInput"
              />

              <!-- Topic suggestions dropdown -->
              <div v-if="showSubjectSuggestions && topicSuggestions.length > 0" class="subject-suggestions">
                <div
                  v-for="(topic, index) in topicSuggestions"
                  :key="topic"
                  class="subject-suggestion"
                  :class="{ 'selected': index === selectedSuggestionIndex }"
                  @click="selectSubjectSuggestion(topic)"
                  @mouseover="selectedSuggestionIndex = index"
                >
                  {{ topic }}
                </div>
              </div>
            </div>
          </div>

          <div class="count-info">
            <span v-if="selectedFilteredCount > 0" class="selected-count">
              Selected: {{ selectedFilteredCount }}
            </span>
            <span>Total: <span id="total-links">{{ filteredLinksCount }}</span></span>
            <div class="settings-dropdown">
              <button
                id="settings-btn"
                title="Settings"
                @click="showSettings = !showSettings"
              >
                <font-awesome-icon icon="cog" />
              </button>
              <div v-if="showSettings" class="settings-menu">
                <div class="settings-header">Compactness</div>
                <button
                  @click="setCompactness('small')"
                  :class="{ active: compactness === 'small' }"
                  class="settings-option"
                >
                  Small
                </button>
                <button
                  @click="setCompactness('medium')"
                  :class="{ active: compactness === 'medium' }"
                  class="settings-option"
                >
                  Medium
                </button>
                <button
                  @click="setCompactness('large')"
                  :class="{ active: compactness === 'large' }"
                  class="settings-option"
                >
                  Large
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="links-container" :class="`compact-${compactness}`">
        <div
            v-for="topic in filteredTopics"
            :key="topic.id"
            class="subject-group"
            :class="{ 'collapsed': isTopicCollapsed(topic.id) }"
          >
            <div class="subject-header" @click="toggleTopicCollapse(topic.id)">
              <div class="collapse-icon">
                {{ isTopicCollapsed(topic.id) ? '▶' : '▼' }}
              </div>
              <div class="subject">{{ topic.name }}</div>
              <div class="link-count">({{ topic.links.length }})</div>
            </div>
            <div class="links-list" v-show="!isTopicCollapsed(topic.id)">
              <div
                v-for="(link, index) in topic.links"
                :key="`${topic.id}-${link.id}`"
                class="link-item"
                :class="{ selected: link.selected, focused: focusedLinkIndex === index }"
                :tabindex="0"
                :data-link-id="link.id"
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
                <span v-if="link.topicName" class="topic-badge">{{ link.topicName }}</span>
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
      :current-subject-id="currentSubjectId"
      :current-subject-name="currentSubjectName"
      :available-topics="availableTopics"
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
import { supabase } from '@/lib/supabase'
import BulkLinkAdder from './components/BulkLinkAdder.vue'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import KeyboardShortcuts from './components/KeyboardShortcuts.vue'
import Login from './components/Login.vue'

export default {
  name: 'App',
  components: {
    BulkLinkAdder,
    FontAwesomeIcon,
    KeyboardShortcuts,
    Login
  },
  data() {
    return {
      isAuthenticated: false,
      user: null,
      subjects: [], // List of subjects (top-level categories)
      currentSubjectId: '', // Currently selected subject ID
      topics: [], // Topics to display (flattened from all subjects or filtered)
      searchQuery: '',
      filteredTopics: [],
      showBulkDialog: false,
      currentTopic: '',
      focusedLinkIndex: -1,
      currentTopicIndex: -1,
      showHelp: false,
      currentTopicFilter: '', // Topic filter (for search like "topic:python")
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
      selectedSuggestionIndex: -1,
      notification: {
        show: false,
        message: '',
        type: 'info',
        timeout: null
      },
      collapsedTopics: new Set(), // Track which topics are collapsed
      compactness: 'medium', // small, medium, large
      showSettings: false,
      isHeaderHidden: false,
      lastScrollTop: 0,
      scrollThreshold: 5, // Minimum scroll distance to trigger hide/show
    }
  },
  computed: {
    hasLinks() {
      return this.topics.some(topic => topic.links && topic.links.length > 0)
    },
    selectedCount() {
      return this.topics.reduce((count, topic) =>
        count + (topic.links ? topic.links.filter(link => link.selected).length : 0), 0)
    },
    totalLinks() {
      return this.topics.reduce((count, topic) =>
        count + (topic.links ? topic.links.length : 0), 0)
    },
    allSelected() {
      return this.hasLinks && this.selectedCount === this.totalLinks
    },
    availableTopics() {
      return this.topics.map(t => t.name).filter(Boolean)
    },
    currentSubjectName() {
      if (!this.currentSubjectId) return ''
      const subject = this.subjects.find(s => s.id == this.currentSubjectId)
      return subject ? subject.name : ''
    },
    // Extract the search part without the topic filter
    searchQueryWithoutTopic() {
      // Remove any existing topic: prefix
      return this.searchQuery.replace(/topic:[^\s]+\s*/, '').trim()
    },
    // Get topic suggestions based on current input
    topicSuggestions() {
      if (!this.showSubjectSuggestions) return []

      // Match topic: prefix
      const match = this.searchQuery.match(/topic:([^&\s]*)/)
      if (!match) return []

      const partialTopic = match[1].toLowerCase()
      if (!partialTopic) return this.availableTopics

      // Filter topics that match the partial input
      return this.availableTopics.filter(topic =>
        topic.toLowerCase().includes(partialTopic)
      )
    },
    // Update to count only filtered links
    filteredLinksCount() {
      return this.filteredTopics.reduce((count, topic) =>
        count + (topic.links ? topic.links.length : 0), 0)
    },

    // Update to check if all filtered links are selected
    allFilteredSelected() {
      if (!this.hasFilteredLinks) return false

      const selectedFilteredCount = this.filteredTopics.reduce((count, topic) =>
        count + (topic.links ? topic.links.filter(link => link.selected).length : 0), 0)

      return selectedFilteredCount === this.filteredLinksCount
    },

    // Add a computed property to check if there are any filtered links
    hasFilteredLinks() {
      return this.filteredTopics.some(topic => topic.links && topic.links.length > 0)
    },
    selectedFilteredCount() {
      return this.filteredTopics.reduce((count, topic) =>
        count + (topic.links ? topic.links.filter(link => link.selected).length : 0), 0)
    }
  },
  async mounted() {
    // Check if user is already logged in
    await this.checkAuth()

    // Listen for auth state changes
    supabase.auth.onAuthStateChange((event, session) => {
      if (event === 'SIGNED_IN' && session) {
        this.isAuthenticated = true
        this.user = session.user
        this.loadSubjects()
      } else if (event === 'SIGNED_OUT') {
        this.isAuthenticated = false
        this.user = null
        this.subjects = []
        this.topics = []
        this.filteredTopics = []
      }
    })

    if (this.isAuthenticated) {
      this.loadSubjects()
    }

    // Load compactness preference from localStorage
    const savedCompactness = localStorage.getItem('compactness')
    if (savedCompactness && ['small', 'medium', 'large'].includes(savedCompactness)) {
      this.compactness = savedCompactness
    }

    // Add global keyboard event listener
    document.addEventListener('keydown', this.handleGlobalKeydown)

    // Add keyboard shortcuts for subject and search focus
    document.addEventListener('keydown', (e) => {
      if (e.ctrlKey && e.key === 'o') {
        e.preventDefault()
        document.getElementById('subject-list').focus()
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

    // Add global shortcut for search
    document.addEventListener('keydown', (e) => {
      // Skip if we're already in an input field
      if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') return

      if (e.key === '/' && !e.ctrlKey && !e.altKey && !e.metaKey) {
        e.preventDefault()
        document.getElementById('search-input').focus()
      }
    })

    // Close subject suggestions when clicking outside
    document.addEventListener('click', this.closeSubjectSuggestions)

    // Close settings menu when clicking outside
    document.addEventListener('click', this.closeSettings)

    // Add scroll listener for auto-hiding header on mobile
    const linksContainer = document.querySelector('.links-container')
    if (linksContainer) {
      linksContainer.addEventListener('scroll', this.handleScroll)
    }

    // Focus first link after loading
    this.$nextTick(() => {
      this.focusFirstLink()
    })
  },
  beforeUnmount() {
    // Clean up event listeners
    document.removeEventListener('keydown', this.handleGlobalKeydown)
    document.removeEventListener('click', this.closeSubjectSuggestions)
    document.removeEventListener('click', this.closeSettings)

    // Clean up scroll listener
    const linksContainer = document.querySelector('.links-container')
    if (linksContainer) {
      linksContainer.removeEventListener('scroll', this.handleScroll)
    }
  },
  methods: {
    handleScroll(e) {
      const scrollTop = e.target.scrollTop

      // Only hide/show header on mobile (check window width)
      if (window.innerWidth > 768) {
        this.isHeaderHidden = false
        return
      }

      // Don't do anything if scroll is very small
      if (Math.abs(scrollTop - this.lastScrollTop) < this.scrollThreshold) {
        return
      }

      // Scrolling down - hide header
      if (scrollTop > this.lastScrollTop && scrollTop > 50) {
        this.isHeaderHidden = true
      }
      // Scrolling up - show header
      else if (scrollTop < this.lastScrollTop) {
        this.isHeaderHidden = false
      }

      this.lastScrollTop = scrollTop
    },
    async checkAuth() {
      const { data: { session } } = await supabase.auth.getSession()

      if (session) {
        this.isAuthenticated = true
        this.user = session.user
        localStorage.setItem('token', session.access_token)
        localStorage.setItem('user', JSON.stringify({
          id: session.user.id,
          email: session.user.email
        }))
      } else {
        this.isAuthenticated = false
        this.user = null
      }
    },
    handleLogin(data) {
      this.isAuthenticated = true
      this.user = data.user
      this.loadSubjects()
      this.showNotification('Welcome back!', 'success')
    },
    async logout() {
      await supabase.auth.signOut()
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      this.isAuthenticated = false
      this.user = null
      this.subjects = []
      this.topics = []
      this.filteredTopics = []
      this.showNotification('Logged out successfully', 'info')
    },
    async loadSubjects() {
      try {
        const { data, error } = await supabase
          .from('subjects')
          .select('*')
          .order('name')

        if (error) throw error

        this.subjects = data || []

        // Restore state from URL parameters (subject and search)
        this.loadFromUrl()

        // Load all topics and links
        await this.loadAllTopicsAndLinks()
      } catch (error) {
        console.error('Error loading subjects:', error)
        this.showNotification('Error loading data', 'error')

        // If JWT error, logout
        if (error.message?.includes('JWT')) {
          this.logout()
        }
      }
    },
    async loadSubjectData() {
      // Load topics and links for the currently selected subject, or all if none selected
      // Clear existing data immediately
      this.topics = []
      this.filteredTopics = []

      // Load new data
      await this.loadAllTopicsAndLinks()

      // Update URL to reflect subject change
      this.updateUrl()
    },
    async loadAllTopicsAndLinks() {
      try {
        this.topics = []

        // Filter subjects based on current selection
        const subjectsToLoad = this.currentSubjectId
          ? this.subjects.filter(s => s.id == this.currentSubjectId)
          : this.subjects

        // Load topics for each subject
        for (const subject of subjectsToLoad) {
          const { data: topics, error: topicsError } = await supabase
            .from('topics')
            .select('*')
            .eq('subject_id', subject.id)
            .order('name')

          if (topicsError) {
            console.error('Failed to load topics for subject:', subject.name, topicsError)
            continue
          }

          // Load links for each topic
          for (const topic of topics || []) {
            const { data: links, error: linksError } = await supabase
              .from('links')
              .select('*')
              .eq('topic_id', topic.id)
              .order('created_at', { ascending: false })

            if (linksError) {
              console.error('Failed to load links for topic:', topic.name, linksError)
              continue
            }

            // Initialize selected property for each link
            (links || []).forEach(link => {
              link.selected = false
            })

            topic.links = links || []
            this.topics.push(topic)
          }
        }

        this.applyFilters()

        // Focus first link after loading
        this.$nextTick(() => {
          this.focusFirstLink()
        })
      } catch (error) {
        console.error('Error loading topics and links:', error)
        this.showNotification('Error loading data', 'error')
      }
    },
    updateUrl() {
      const newUrl = new URL(window.location)

      // Update subject parameter - use subject NAME not ID
      if (this.currentSubjectId) {
        const subject = this.subjects.find(s => s.id == this.currentSubjectId)
        if (subject) {
          newUrl.searchParams.set('subject', subject.name)
        }
      } else {
        newUrl.searchParams.delete('subject')
      }

      // Update search parameter
      if (this.searchQuery) {
        newUrl.searchParams.set('search', this.searchQuery)
      } else {
        newUrl.searchParams.delete('search')
      }

      window.history.pushState({}, '', newUrl)
    },
    loadFromUrl() {
      // Parse URL parameters and restore state
      const urlParams = new URLSearchParams(window.location.search)

      // Restore subject selection
      const subjectName = urlParams.get('subject')
      if (subjectName) {
        const subject = this.subjects.find(s => s.name.toLowerCase() === subjectName.toLowerCase())
        if (subject) {
          this.currentSubjectId = subject.id
        }
      }

      // Restore search query
      const search = urlParams.get('search')
      if (search) {
        this.searchQuery = search
      }
    },
    applyFilters() {
      // Parse topic filter from search query
      const topicMatch = this.searchQuery.match(/topic:([^\s]+)/)
      if (topicMatch) {
        this.currentTopicFilter = topicMatch[1]
      } else {
        this.currentTopicFilter = ''
      }

      // Get search terms (everything after the topic: part if it exists)
      let searchTerms = ''
      if (topicMatch) {
        // Extract everything after the topic: part
        const afterTopic = this.searchQuery.substring(
          this.searchQuery.indexOf(topicMatch[0]) + topicMatch[0].length
        ).trim()
        searchTerms = afterTopic
      } else {
        searchTerms = this.searchQuery.trim()
      }

      // Apply filters
      this.filterLinks(this.currentTopicFilter, searchTerms)

      // Update URL
      this.updateUrl()
    },
    filterLinks(topicFilter, searchTerms) {
      // First filter by topic if needed
      let filteredByTopic = this.topics
      if (topicFilter) {
        filteredByTopic = this.topics.filter(topic =>
          topic.name.toLowerCase() === topicFilter.toLowerCase()
        )
      }

      // If no search terms after filtering by topic, return all topics
      // Force new array to trigger Vue reactivity
      if (!searchTerms) {
        this.filteredTopics = filteredByTopic.map(topic => ({
          ...topic,
          links: [...(topic.links || [])]
        }))
        return
      }

      // Parse search terms for OR/AND operators
      const { andGroups, orTerms } = this.parseSearchTerms(searchTerms)

      // When searching, flatten all links and sort by relevance globally
      const allRelevantLinks = []

      filteredByTopic.forEach(topic => {
        topic.links.forEach(link => {
          const relevance = this.calculateRelevance(link, andGroups, orTerms)
          if (relevance > 0) {
            allRelevantLinks.push({
              ...link,
              relevance,
              topicName: topic.name,
              topicId: topic.id
            })
          }
        })
      })

      // Sort all links by relevance (highest first)
      allRelevantLinks.sort((a, b) => b.relevance - a.relevance)

      // Create a single "Search Results" topic containing all sorted links
      if (allRelevantLinks.length > 0) {
        this.filteredTopics = [{
          id: 'search-results',
          name: 'Search Results',
          links: allRelevantLinks
        }]
      } else {
        this.filteredTopics = []
      }
    },
    parseSearchTerms(searchTerms) {
      // Split by OR first
      const orParts = searchTerms.split(/\s+OR\s+|\s+\|\s+/).map(p => p.trim())

      const andGroups = []
      const orTerms = []

      orParts.forEach(part => {
        if (part.includes(' AND ') || part.includes(' & ')) {
          // This is an AND group
          const terms = part.split(/\s+AND\s+|\s+&\s+/)
            .map(t => t.trim().toLowerCase())
            .filter(t => t.length > 0)
          if (terms.length > 0) {
            andGroups.push(terms)
          }
        } else {
          // Regular OR terms (space-separated words)
          const terms = part.split(/\s+/)
            .map(t => t.trim().toLowerCase())
            .filter(t => t.length > 0 && t !== 'or' && t !== 'and' && t !== '&' && t !== '|')
          orTerms.push(...terms)
        }
      })

      return { andGroups, orTerms }
    },
    calculateRelevance(link, andGroups, orTerms) {
      const titleLower = link.title.toLowerCase()
      const urlLower = link.url.toLowerCase()
      const titleWords = titleLower.split(/\s+/)
      const urlWords = urlLower.split(/\s+/)

      let relevance = 0

      // Check AND groups - ALL terms in at least one group must match
      let andGroupMatched = andGroups.length === 0 // If no AND groups, consider it matched
      for (const andGroup of andGroups) {
        let allTermsInGroupMatched = true
        let groupRelevance = 0

        for (const term of andGroup) {
          const termMatch = this.matchTerm(term, titleLower, titleWords, urlLower, urlWords)
          if (termMatch > 0) {
            groupRelevance += termMatch
          } else {
            allTermsInGroupMatched = false
            break
          }
        }

        if (allTermsInGroupMatched) {
          andGroupMatched = true
          relevance += groupRelevance * 1.5 // Boost AND matches
          break // Only need one AND group to match
        }
      }

      // If AND groups exist but none matched, return 0
      if (andGroups.length > 0 && !andGroupMatched) {
        return 0
      }

      // Check OR terms - ANY term can match
      let orMatchCount = 0
      for (const term of orTerms) {
        const termMatch = this.matchTerm(term, titleLower, titleWords, urlLower, urlWords)
        if (termMatch > 0) {
          relevance += termMatch
          orMatchCount++
        }
      }

      // If we have OR terms but none matched, return 0
      if (orTerms.length > 0 && orMatchCount === 0 && andGroups.length === 0) {
        return 0
      }

      // Boost if multiple OR terms matched
      if (orMatchCount > 1) {
        relevance *= (1 + (orMatchCount * 0.2))
      }

      return relevance
    },
    matchTerm(term, titleLower, titleWords, urlLower, urlWords) {
      let score = 0

      // Exact match in title
      if (titleLower === term) {
        score += 5
      }
      // Title contains the exact term
      else if (titleLower.includes(term)) {
        score += 3
      }

      // Check individual title words
      for (const word of titleWords) {
        if (word === term) {
          score += 2
          break
        } else if (word.includes(term)) {
          score += 1
          break
        }
      }

      // Check URL
      if (urlLower.includes(term)) {
        score += 1
      }

      // Check individual URL words
      for (const word of urlWords) {
        if (word === term) {
          score += 0.8
          break
        } else if (word.includes(term)) {
          score += 0.4
          break
        }
      }

      return score
    },
    onSearchInput: debounce(function() {
      // Check if we're typing in the topic: part
      if (this.searchQuery.match(/topic:[^&\s]*$/)) {
        this.showSubjectSuggestions = true
        this.selectedSuggestionIndex = -1 // Reset selection index when input changes
      } else {
        this.showSubjectSuggestions = false
      }

      this.applyFilters()
    }, 300),
    handleSearchKeydown(e) {
      // Handle navigation in topic suggestions
      if (this.showSubjectSuggestions && this.topicSuggestions.length > 0) {
        if (e.key === 'ArrowDown') {
          e.preventDefault()
          this.selectedSuggestionIndex = Math.min(
            this.selectedSuggestionIndex + 1,
            this.topicSuggestions.length - 1
          )
          if (this.selectedSuggestionIndex === -1) {
            this.selectedSuggestionIndex = 0
          }
        } else if (e.key === 'ArrowUp') {
          e.preventDefault()
          this.selectedSuggestionIndex = Math.max(this.selectedSuggestionIndex - 1, 0)
        } else if (e.key === 'Enter' || e.key === 'Tab') {
          e.preventDefault()
          if (this.topicSuggestions.length > 0) {
            // If a suggestion is highlighted, select it
            if (this.selectedSuggestionIndex >= 0) {
              this.selectSubjectSuggestion(this.topicSuggestions[this.selectedSuggestionIndex])
            } else {
              // Otherwise select the first suggestion
              this.selectSubjectSuggestion(this.topicSuggestions[0])
            }
          }
        } else if (e.key === 'Escape') {
          e.preventDefault()
          this.showSubjectSuggestions = false
          this.selectedSuggestionIndex = -1
        }
      }
    },
    selectSubjectSuggestion(topic) {
      // Replace the partial topic with the selected one
      const beforeTopic = this.searchQuery.split('topic:')[0]
      let afterTopic = ''

      // Check if there's content after the topic
      const afterMatch = this.searchQuery.match(/topic:[^\s]*(.*?)$/)
      if (afterMatch && afterMatch[1]) {
        afterTopic = afterMatch[1]
      }

      this.searchQuery = `${beforeTopic}topic:${topic}${afterTopic}`
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
    closeSettings(event) {
      // Don't close if clicking inside the settings dropdown
      const settingsDropdown = event.target.closest('.settings-dropdown')
      if (settingsDropdown) {
        return
      }

      this.showSettings = false
    },
    async deleteSelected() {
      // Get all visible links from filtered topics
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
        const linkIds = selectedLinks.map(link => link.id)

        console.log('Attempting to delete links:', linkIds)

        const { data, error } = await supabase
          .from('links')
          .delete()
          .in('id', linkIds)
          .select()

        if (error) {
          console.error('Delete error details:', error)
          throw error
        }

        console.log('Delete successful, deleted:', data)
        this.showNotification(`Deleted ${selectedLinks.length} links`, 'success')
        // Reload links after deletion
        await this.loadAllTopicsAndLinks()
      } catch (error) {
        this.showNotification(`Error deleting links: ${error.message}`, 'error')
        console.error('Error deleting links:', error)
      }
    },
    toggleAllFiltered() {
      const shouldSelect = !this.allFilteredSelected

      // Directly update the filtered topics
      this.filteredTopics.forEach(topic => {
        topic.links.forEach(link => {
          // Find the original link in the topics array and update it
          for (const originalTopic of this.topics) {
            const originalLink = originalTopic.links.find(l => l.id === link.id)
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
        .map(link => `- [${link.title}](${link.url})  `)
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
      // Create a map of IDs for faster lookup
      const visibleLinks = this.getVisibleLinks()
      const visibleIdMap = new Set(visibleLinks.map(link => link.id))

      // Only update if the link is visible in the filtered view
      if (visibleIdMap.has(changedLink.id)) {
        // Update all instances of this link in both original and filtered data
        this.topics.forEach(topic => {
          topic.links.forEach(link => {
            if (link.id === changedLink.id) {
              link.selected = changedLink.selected
            }
          })
        })

        this.filteredTopics.forEach(topic => {
          topic.links.forEach(link => {
            if (link.id === changedLink.id) {
              link.selected = changedLink.selected
            }
          })
        })
      }
    },
    openBulkAddDialog(topicName) {
      this.currentTopic = topicName
      this.showBulkDialog = true
    },
    async handleLinksAdded() {
      this.showBulkDialog = false
      await this.loadAllTopicsAndLinks()
    },
    setCompactness(level) {
      this.compactness = level
      localStorage.setItem('compactness', level)
      this.showSettings = false
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
      const allLinks = this.getAllLinks()
      let currentIndex = allLinks.indexOf(focusedElement)

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
          // If no link is focused, focus the first one
          if (currentIndex === -1 && allLinks.length > 0) {
            allLinks[0].focus()
            currentIndex = 0
          }

          if (currentIndex >= 0 && allLinks[currentIndex]) {
            const linkElement = allLinks[currentIndex]
            const linkId = linkElement.getAttribute('data-link-id')

            // Search through filtered topics first for better performance
            let found = false
            for (const topic of this.filteredTopics) {
              for (const link of topic.links) {
                if (link.id === linkId) {
                  link.selected = !link.selected
                  this.onCheckboxChange(link)
                  found = true
                  break
                }
              }
              if (found) break
            }

            // If not found in filtered topics, search all topics
            if (!found) {
              for (const topic of this.topics) {
                for (const link of topic.links) {
                  if (link.id === linkId) {
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
          // If no link is focused, focus the first one
          if (currentIndex === -1 && allLinks.length > 0) {
            allLinks[0].focus()
          } else if (currentIndex >= 0 && allLinks[currentIndex]) {
            const linkId = allLinks[currentIndex].getAttribute('data-link-id')
            const link = this.getVisibleLinks().find(l => l.id === linkId)
            if (link) {
              this.openLink(link)
            }
          }
          break
        case 'ArrowDown':
        case 'j':
          e.preventDefault()
          // If no link is focused, focus the first one
          if (currentIndex === -1 && allLinks.length > 0) {
            allLinks[0].focus()
          } else {
            this.focusNextLink(currentIndex)
          }
          break
        case 'ArrowUp':
        case 'k':
          e.preventDefault()
          // If no link is focused, focus the first or last one
          if (currentIndex === -1 && allLinks.length > 0) {
            allLinks[allLinks.length - 1].focus()
          } else {
            this.focusPreviousLink(currentIndex)
          }
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
      // First get all links from filtered topics
      return this.filteredTopics.flatMap(topic => topic.links || [])
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
    // Toggle collapse state for a topic
    toggleTopicCollapse(topicId) {
      if (this.collapsedTopics.has(topicId)) {
        this.collapsedTopics.delete(topicId)
      } else {
        this.collapsedTopics.add(topicId)
      }
    },

    // Check if a topic is collapsed
    isTopicCollapsed(topicId) {
      return this.collapsedTopics.has(topicId)
    },
    
    // Update getAllLinks to only include links from expanded subjects
    getAllLinks() {
      return Array.from(document.querySelectorAll('.link-item'))
        .filter(link => {
          const style = window.getComputedStyle(link)
          return style.display !== 'none' && style.visibility !== 'hidden'
        })
    },
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
  margin: 0;
  padding: 0;
  overflow: hidden;
}

.app {
  padding: 0;
  width: 100vw;
  height: 100vh;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-container {
  width: 100%;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 0 20px 0 20px;
}

.app-header {
  padding: 8px 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: var(--card-background);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
  box-sizing: border-box;
  transition: transform 0.3s ease, opacity 0.3s ease;
  position: sticky;
  top: 0;
  z-index: 100;
}

.app-title {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--primary-color);
}

.action-controls {
  display: flex;
  align-items: center;
  gap: 20px;
}

.search-controls {
  display: flex;
  gap: 12px;
  align-items: center;
  flex: 1;
  max-width: 700px;
}

#subject-list {
  flex: 0 0 200px;
  min-width: 150px;
  padding: 9px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background-color: var(--card-background);
  color: var(--text-color);
  transition: all 0.2s;
  cursor: pointer;
}

.search-container {
  position: relative;
  flex: 1;
  box-sizing: border-box;
}

#search-input {
  width: 100%;
  padding: 9px 12px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background-color: var(--card-background);
  color: var(--text-color);
  transition: all 0.2s;
  box-sizing: border-box;
}

#subject-list:hover,
#search-input:hover {
  border-color: #9ca3af;
}

#subject-list:focus,
#search-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
}

.multi-select-controls {
  padding: 12px 20px;
  background-color: var(--card-background);
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  box-sizing: border-box;
  transition: transform 0.3s ease, opacity 0.3s ease;
  position: sticky;
  top: 0;
  z-index: 99;
}

.links-container {
  flex: 1;
  width: 100%;
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 20px 20px 20px;
  box-sizing: border-box;
  overflow-y: auto;
  overflow-x: hidden;
  background-color: var(--card-background);
}

.count-info {
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: auto;
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

.settings-dropdown {
  position: relative;
  display: flex;
  align-items: center;
}

#settings-btn {
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

#settings-btn:hover {
  background-color: #f0f0f0;
  color: var(--primary-color);
}

.settings-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 5px;
  background: white;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
  min-width: 140px;
  padding: 8px 0;
}

.settings-header {
  padding: 6px 12px;
  font-size: 11px;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 4px;
}

.settings-option {
  display: block;
  width: 100%;
  padding: 8px 16px;
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-color);
  transition: background-color 0.2s;
}

.settings-option:hover {
  background-color: rgba(52, 152, 219, 0.1);
}

.settings-option.active {
  background-color: rgba(52, 152, 219, 0.15);
  color: var(--primary-color);
  font-weight: 500;
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
  margin-bottom: 6px;
}

.link-item {
  display: flex;
  align-items: center;
  padding: 2px 8px;
  border-left: 3px solid transparent;
  overflow: hidden;
  width: 100%;
  outline: none;
  transition: all 0.15s ease;
  border-radius: 4px;
  margin-bottom: 1px;
}

.link-item p {
  margin: 0;
  flex-grow: 1;
  font-size: 13px;
  line-height: 1.2;
  padding-left: 6px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.topic-badge {
  background-color: #e3f2fd;
  color: #1976d2;
  padding: 1px 6px;
  border-radius: 8px;
  font-size: 10px;
  font-weight: 500;
  white-space: nowrap;
  margin-left: auto;
  flex-shrink: 0;
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
  width: 18px;
}

.checkbox-wrapper input[type="checkbox"] {
  width: 14px;
  height: 14px;
  margin: 0;
  cursor: pointer;
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
  margin-bottom: 12px;
  padding-bottom: 0;
}

.subject-group:last-child {
  margin-bottom: 0;
}

.subject-group.collapsed {
  margin-bottom: 6px;
}

/* Logout button */
.logout-btn {
  padding: 9px 16px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 13px;
  background-color: var(--card-background);
  color: var(--text-color);
  cursor: pointer;
  transition: all 0.2s;
  font-weight: 500;
  white-space: nowrap;
}

.logout-btn:hover {
  background-color: #f3f4f6;
  border-color: #9ca3af;
}

.logout-btn:active {
  transform: scale(0.98);
}

/* Compactness styles */

/* Small compactness - tight spacing */
.compact-small .checkbox-wrapper {
  width: 18px;
}

.compact-small .checkbox-wrapper input[type="checkbox"] {
  width: 14px;
  height: 14px;
}

.compact-small .link-item {
  padding: 2px 8px;
  margin-bottom: 1px;
}

.compact-small .link-item p {
  font-size: 13px;
  line-height: 1.2;
  padding-left: 6px;
}

.compact-small .links-list {
  margin-bottom: 6px;
}

.compact-small .subject-group {
  margin-bottom: 12px;
}

.compact-small .subject-group.collapsed {
  margin-bottom: 6px;
}

/* Medium compactness - default spacing */
.compact-medium .checkbox-wrapper {
  width: 25px;
}

.compact-medium .checkbox-wrapper input[type="checkbox"] {
  width: 16px;
  height: 16px;
}

.compact-medium .link-item {
  padding: 4px 8px;
  margin-bottom: 2px;
}

.compact-medium .link-item p {
  font-size: 13px;
  line-height: 1.3;
  padding-left: 8px;
}

.compact-medium .links-list {
  margin-bottom: 10px;
}

.compact-medium .subject-group {
  margin-bottom: 16px;
}

.compact-medium .subject-group.collapsed {
  margin-bottom: 8px;
}

/* Large compactness - spacious */
.compact-large .checkbox-wrapper {
  width: 30px;
}

.compact-large .checkbox-wrapper input[type="checkbox"] {
  width: 18px;
  height: 18px;
}

.compact-large .link-item {
  padding: 6px 10px;
  margin-bottom: 3px;
}

.compact-large .link-item p {
  font-size: 14px;
  line-height: 1.4;
  padding-left: 10px;
}

.compact-large .links-list {
  margin-bottom: 14px;
}

.compact-large .subject-group {
  margin-bottom: 20px;
}

.compact-large .subject-group.collapsed {
  margin-bottom: 10px;
}

/* Header hide/show animation */
.app-header.header-hidden {
  transform: translateY(-100%);
  opacity: 0;
}

.multi-select-controls.controls-hidden {
  transform: translateY(-100%);
  opacity: 0;
}

/* Mobile responsive styles */
@media (max-width: 768px) {
  body {
    font-size: 16px; /* Prevent iOS zoom on input focus */
  }

  /* Adjust sticky positioning for mobile */
  .app-header {
    top: 0;
  }

  .multi-select-controls {
    top: 0;
  }

  /* When header is hidden, adjust controls position */
  .app-header.header-hidden + .multi-select-controls:not(.controls-hidden) {
    top: 0;
  }

  .app-container {
    padding: 0 10px 0 10px;
  }

  .app-header {
    padding: 10px;
    flex-direction: row;
    gap: 10px;
  }

  .app-title {
    font-size: 16px;
  }

  .logout-btn {
    padding: 8px 12px;
    font-size: 12px;
  }

  .multi-select-controls {
    padding: 10px;
  }

  .action-controls {
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }

  .control-buttons {
    width: 100%;
    justify-content: space-between;
    order: 3; /* Move to bottom */
  }

  .search-controls {
    flex-direction: column;
    max-width: 100%;
    width: 100%;
    order: 2;
    gap: 8px;
  }

  #subject-list {
    flex: none;
    width: 100%;
    padding: 12px;
    font-size: 16px; /* Prevent iOS zoom */
  }

  .search-container {
    width: 100%;
  }

  #search-input {
    width: 100%;
    padding: 12px;
    font-size: 16px; /* Prevent iOS zoom */
  }

  .count-info {
    width: 100%;
    justify-content: space-between;
    order: 1; /* Move to top */
    margin-left: 0;
    font-size: 14px;
  }

  .control-buttons {
    flex-wrap: nowrap;
    overflow-x: visible;
  }

  .control-buttons button {
    height: 36px;
    width: 36px;
    font-size: 16px;
    flex-shrink: 0;
  }

  .control-buttons button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
    display: inline-flex !important;
  }

  .control-buttons input[type="checkbox"] {
    width: 18px;
    height: 18px;
  }

  .links-container {
    padding: 0 10px 20px 10px;
  }

  .link-item {
    padding: 8px;
    margin-bottom: 2px;
  }

  .link-item p {
    font-size: 14px;
    line-height: 1.4;
  }

  .checkbox-wrapper input[type="checkbox"] {
    width: 18px;
    height: 18px;
  }

  .subject-header {
    padding: 8px;
    margin-bottom: 8px;
  }

  .subject {
    font-size: 1.1em;
  }

  .help-button {
    width: 50px;
    height: 50px;
    font-size: 24px;
    bottom: 15px;
    right: 15px;
  }

  .notification {
    bottom: 75px;
    right: 15px;
    left: 15px;
    max-width: none;
  }

  .settings-dropdown {
    position: relative;
  }

  #settings-btn {
    height: 36px;
    width: 36px;
    font-size: 16px;
  }

  .settings-menu {
    right: -10px;
  }

  /* Make topic badges smaller on mobile */
  .topic-badge {
    font-size: 9px;
    padding: 2px 4px;
  }

  /* Adjust subject suggestions for mobile */
  .subject-suggestions {
    max-height: 150px;
  }

  .subject-suggestion {
    padding: 12px;
    font-size: 16px;
  }

  /* Compact mode overrides for mobile */
  .compact-small .link-item,
  .compact-medium .link-item,
  .compact-large .link-item {
    padding: 8px;
    margin-bottom: 2px;
  }

  .compact-small .link-item p,
  .compact-medium .link-item p,
  .compact-large .link-item p {
    font-size: 14px;
    line-height: 1.4;
  }

  .compact-small .checkbox-wrapper,
  .compact-medium .checkbox-wrapper,
  .compact-large .checkbox-wrapper {
    width: 28px;
  }

  .compact-small .checkbox-wrapper input[type="checkbox"],
  .compact-medium .checkbox-wrapper input[type="checkbox"],
  .compact-large .checkbox-wrapper input[type="checkbox"] {
    width: 18px;
    height: 18px;
  }
}

/* Extra small devices (phones in portrait, less than 576px) */
@media (max-width: 576px) {
  .app-title {
    font-size: 14px;
  }

  .logout-btn {
    padding: 6px 10px;
    font-size: 11px;
  }

  .link-item p {
    font-size: 13px;
  }

  .count-info {
    font-size: 12px;
    flex-wrap: wrap;
  }

  .selected-count {
    margin-right: 8px;
  }
}
</style>
