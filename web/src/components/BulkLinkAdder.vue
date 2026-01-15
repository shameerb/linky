<template>
  <div class="bulk-link-dialog" @click.self="$emit('close')">
    <div class="dialog-content">
      <button class="close-btn" @click="$emit('close')">×</button>
      <h3>Add Bulk Links</h3>

      <div class="subject-select">
        <label>Topic</label>
        <div class="autocomplete-wrapper">
          <input
            v-model="topicName"
            type="text"
            class="subject-input"
            placeholder="Enter or select topic name"
            @input="onTopicInput"
            @keydown="handleTopicKeydown"
            ref="topicInput"
          />
          <div v-if="showSuggestions && filteredTopics.length > 0" class="suggestions-dropdown">
            <div
              v-for="(topic, index) in filteredTopics"
              :key="topic"
              class="suggestion-item"
              :class="{ 'selected': index === selectedIndex }"
              @click="selectTopic(topic)"
              @mouseover="selectedIndex = index"
            >
              {{ topic }}
            </div>
          </div>
        </div>
      </div>

      <p class="help-text">Enter links in markdown format:</p>
      <pre class="format-example">- [link title](link url)
- [another title](another url)  </pre>
      <textarea
        v-model="linksText"
        placeholder="Paste your markdown links here..."
        rows="10"
      ></textarea>
      <div class="dialog-actions">
        <button @click="$emit('close')" class="cancel-btn">Cancel</button>
        <button @click="addLinks" class="add-btn" :disabled="!canAddLinks">
          Add Links
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { supabase } from '@/lib/supabase'

export default {
  name: 'BulkLinkAdder',
  props: {
    currentSubjectId: {
      type: String,
      default: ''
    },
    currentSubjectName: {
      type: String,
      default: ''
    },
    availableTopics: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      linksText: '',
      topicName: '',
      showSuggestions: false,
      selectedIndex: -1
    }
  },
  computed: {
    canAddLinks() {
      return this.linksText.trim() && this.topicName.trim() && this.currentSubjectId
    },
    filteredTopics() {
      if (!this.topicName) return this.availableTopics
      const query = this.topicName.toLowerCase()
      return this.availableTopics.filter(topic =>
        topic.toLowerCase().includes(query)
      )
    }
  },
  mounted() {
    document.addEventListener('keydown', this.handleEscape)
  },
  beforeUnmount() {
    document.removeEventListener('keydown', this.handleEscape)
  },
  methods: {
    handleEscape(e) {
      if (e.key === 'Escape') {
        e.preventDefault()
        this.$emit('close')
      }
    },
    onTopicInput() {
      this.showSuggestions = true
      this.selectedIndex = -1
    },
    handleTopicKeydown(e) {
      if (!this.showSuggestions || this.filteredTopics.length === 0) return

      if (e.key === 'ArrowDown') {
        e.preventDefault()
        this.selectedIndex = Math.min(
          this.selectedIndex + 1,
          this.filteredTopics.length - 1
        )
        if (this.selectedIndex === -1) {
          this.selectedIndex = 0
        }
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        this.selectedIndex = Math.max(this.selectedIndex - 1, 0)
      } else if (e.key === 'Enter' && this.selectedIndex >= 0) {
        e.preventDefault()
        this.selectTopic(this.filteredTopics[this.selectedIndex])
      } else if (e.key === 'Escape') {
        e.preventDefault()
        this.showSuggestions = false
        this.selectedIndex = -1
      } else if (e.key === 'Tab' && this.filteredTopics.length > 0) {
        e.preventDefault()
        const topicToSelect = this.selectedIndex >= 0
          ? this.filteredTopics[this.selectedIndex]
          : this.filteredTopics[0]
        this.selectTopic(topicToSelect)
      }
    },
    selectTopic(topic) {
      this.topicName = topic
      this.showSuggestions = false
      this.selectedIndex = -1
    },
    parseLinks(text) {
      const links = []
      const lines = text.split('\n')
      const linkRegex = /\[(.*?)\]\((.*?)\)/

      for (const line of lines) {
        if (line.trim()) {
          const matches = line.match(linkRegex)
          if (matches && matches.length === 3) {
            links.push({
              title: matches[1].trim(),
              url: matches[2].trim()
            })
          }
        }
      }
      return links
    },

    async addLinks() {
      if (!this.canAddLinks) return

      const links = this.parseLinks(this.linksText)
      if (links.length === 0) {
        alert('No valid links found. Please check the format.')
        return
      }

      try {
        const subjectId = this.currentSubjectId

        // Check if topic exists
        const { data: topics, error: topicsError } = await supabase
          .from('topics')
          .select('*')
          .eq('subject_id', subjectId)
          .ilike('name', this.topicName)
          .limit(1)

        if (topicsError) throw topicsError

        let topicId = null

        if (topics && topics.length > 0) {
          // Use existing topic
          topicId = topics[0].id
        } else {
          // Create new topic
          const { data: newTopic, error: createError } = await supabase
            .from('topics')
            .insert({
              subject_id: subjectId,
              name: this.topicName
            })
            .select()
            .single()

          if (createError) throw createError
          topicId = newTopic.id
        }

        // Bulk insert links
        const linksToInsert = links.map(link => ({
          topic_id: topicId,
          title: link.title,
          url: link.url
        }))

        const { error: linksError } = await supabase
          .from('links')
          .insert(linksToInsert)

        if (linksError) throw linksError

        this.$emit('links-added')
        this.$emit('close')
        this.linksText = ''
        this.topicName = ''
      } catch (error) {
        console.error('Error adding links:', error)
        alert(`Failed to add links: ${error.message}`)
      }
    }
  }
}
</script>

<style scoped>
.bulk-link-dialog {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog-content {
  background: white;
  padding: 30px;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
  position: relative;
}

.close-btn {
  position: absolute;
  top: 10px;
  right: 15px;
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: #666;
  width: 30px;
  height: 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.close-btn:hover {
  background: #f0f0f0;
  color: #333;
}

.dialog-content h3 {
  margin: 0 0 15px;
  color: var(--secondary-color);
}

.help-text {
  margin: 0 0 5px;
  color: #666;
  font-size: 0.9em;
}

.format-example {
  background: #f5f5f5;
  padding: 10px;
  border-radius: 4px;
  margin: 0 0 15px;
  font-size: 0.9em;
  color: #666;
  white-space: pre;
}

textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 14px;
  font-family: monospace;
  resize: vertical;
  margin-bottom: 15px;
  box-sizing: border-box;
}

textarea:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.dialog-actions button {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  border: none;
  transition: all 0.2s;
}

.cancel-btn {
  background: #f5f5f5;
  color: #666;
}

.cancel-btn:hover {
  background: #e0e0e0;
}

.add-btn {
  background: var(--primary-color);
  color: white;
}

.add-btn:hover {
  background: #2980b9;
}

.add-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}

.subject-select {
  margin-bottom: 15px;
}

.subject-select label {
  display: block;
  margin-bottom: 5px;
  color: #666;
  font-size: 0.9em;
}

.autocomplete-wrapper {
  position: relative;
}

.subject-input {
  width: 100%;
  padding: 8px;
  border: 1px solid var(--border-color);
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.subject-input:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.suggestions-dropdown {
  position: absolute;
  top: 100%;
  left: 0;
  right: 0;
  background: white;
  border: 1px solid var(--border-color);
  border-top: none;
  border-radius: 0 0 4px 4px;
  max-height: 200px;
  overflow-y: auto;
  z-index: 10;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.suggestion-item {
  padding: 8px 12px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.suggestion-item:hover,
.suggestion-item.selected {
  background-color: rgba(52, 152, 219, 0.1);
}
</style>
