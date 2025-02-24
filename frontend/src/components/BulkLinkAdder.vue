<template>
  <div class="bulk-link-dialog">
    <div class="dialog-content">
      <h3>Add Bulk Links</h3>
      <p class="help-text">Enter links in markdown format:</p>
      <pre class="format-example">- [link title](link url)
- [another title](another url)</pre>
      <textarea
        v-model="linksText"
        placeholder="Paste your markdown links here..."
        rows="10"
      ></textarea>
      <div class="dialog-actions">
        <button @click="$emit('close')" class="cancel-btn">Cancel</button>
        <button @click="addLinks" class="add-btn" :disabled="!linksText.trim()">
          Add Links
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'BulkLinkAdder',
  props: {
    currentFile: {
      type: String,
      required: true
    },
    currentSubject: {
      type: String,
      required: true
    }
  },
  data() {
    return {
      linksText: ''
    }
  },
  methods: {
    async addLinks() {
      if (!this.linksText.trim()) return

      console.log('Adding bulk links:', {
        filename: this.currentFile,
        subject: this.currentSubject,
        contentLength: this.linksText.length
      })

      try {
        const response = await fetch('/api/bulk_links', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            filename: this.currentFile,
            content: this.linksText
          })
        })

        console.log('Response status:', response.status)

        if (!response.ok) {
          const errorText = await response.text()
          console.error('Error response:', errorText)
          throw new Error(`Failed to add links: ${errorText}`)
        }

        this.$emit('links-added')
        this.$emit('close')
        this.linksText = ''
      } catch (error) {
        console.error('Error adding links:', error)
        alert('Failed to add links. Please try again.')
      }
    }
  }
}
</script>

<style>
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
  padding: 20px;
  border-radius: 8px;
  width: 90%;
  max-width: 600px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
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
</style>
