<template>
  <div class="help-overlay" v-if="show" @click="close">
    <div class="help-dialog" @click.stop>
      <div class="help-header">
        <h2>Help</h2>
        <button class="close-btn" @click="close" title="Close">×</button>
      </div>
      <div class="help-content">
        <div class="help-columns">
          <div class="help-section">
            <h3>Keyboard Shortcuts</h3>
            <div class="shortcut-group">
              <h4>Navigation</h4>
              <div class="shortcut-item">
                <kbd>j</kbd> or <kbd>↓</kbd>
                <span>Next link</span>
              </div>
              <div class="shortcut-item">
                <kbd>k</kbd> or <kbd>↑</kbd>
                <span>Previous link</span>
              </div>
            </div>
            <div class="shortcut-group">
              <h4>Selection</h4>
              <div class="shortcut-item">
                <kbd>x</kbd>
                <span>Select/deselect link</span>
              </div>
            </div>
            <div class="shortcut-group">
              <h4>Actions</h4>
              <div class="shortcut-item">
                <kbd>Enter</kbd>
                <span>Open focused link</span>
              </div>
              <div class="shortcut-item">
                <kbd>Shift</kbd> + <kbd>|</kbd>
                <span>Open selected links</span>
              </div>
              <div class="shortcut-item">
                <kbd>Shift</kbd> + <kbd>C</kbd>
                <span>Copy selected links</span>
              </div>
              <div class="shortcut-item">
                <kbd>Shift</kbd> + <kbd>#</kbd>
                <span>Delete selected links</span>
              </div>
              <div class="shortcut-item">
                <kbd>Shift</kbd> + <kbd>N</kbd>
                <span>Add new links</span>
              </div>
            </div>
          </div>

          <div class="help-section">
            <h3>Focus Shortcuts</h3>
            <div class="shortcut-group">
              <div class="shortcut-item">
                <kbd>/</kbd>
                <span>Focus search</span>
              </div>
              <div class="shortcut-item">
                <kbd>Ctrl</kbd> + <kbd>F</kbd>
                <span>Focus search</span>
              </div>
              <div class="shortcut-item">
                <kbd>Ctrl</kbd> + <kbd>O</kbd>
                <span>Focus subject selector</span>
              </div>
              <div class="shortcut-item">
                <kbd>?</kbd>
                <span>Toggle this help</span>
              </div>
            </div>
          </div>
        </div>

        <div class="help-section">
          <h3>Search Queries</h3>
          <div class="shortcut-group">
            <div class="shortcut-item">
              <code>python</code>
              <span>Search for "python"</span>
            </div>
            <div class="shortcut-item">
              <code>topic:ai</code>
              <span>Filter to topic</span>
            </div>
            <div class="shortcut-item">
              <code>machine AND learning</code>
              <span>Both terms required</span>
            </div>
            <div class="shortcut-item">
              <code>javascript OR typescript</code>
              <span>Either term</span>
            </div>
            <div class="shortcut-item">
              <code>deep & learning</code>
              <span>AND (alternative: &)</span>
            </div>
            <div class="shortcut-item">
              <code>react | vue</code>
              <span>OR (alternative: |)</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'KeyboardShortcuts',
  props: {
    show: {
      type: Boolean,
      default: false
    }
  },
  mounted() {
    document.addEventListener('keydown', this.handleKeydown)
  },
  beforeUnmount() {
    document.removeEventListener('keydown', this.handleKeydown)
  },
  methods: {
    close() {
      this.$emit('close')
    },
    handleKeydown(e) {
      if (e.key === 'Escape' && this.show) {
        e.preventDefault()
        this.close()
      }
    }
  }
}
</script>

<style scoped>
.help-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.help-dialog {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  border-radius: 12px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.2);
  width: 90%;
  max-width: 800px;
  max-height: 85vh;
  overflow-y: auto;
  animation: slideIn 0.2s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.help-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 18px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

.help-header h2 {
  margin: 0;
  color: #333;
  font-size: 18px;
}

.help-content {
  padding: 16px 18px;
}

.help-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 16px;
}

.help-section {
  margin-bottom: 16px;
}

.help-section h3 {
  margin: 0 0 10px 0;
  color: #3498db;
  font-size: 14px;
  font-weight: 600;
}

.shortcut-group {
  margin-bottom: 12px;
}

.shortcut-group h4 {
  margin: 0 0 6px 0;
  color: #666;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.shortcut-item {
  display: flex;
  align-items: center;
  margin-bottom: 6px;
  font-size: 12px;
}

kbd {
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 3px;
  padding: 2px 6px;
  box-shadow: 0 1px 1px rgba(0,0,0,0.1);
  font-family: monospace;
  font-size: 11px;
  min-width: 20px;
  text-align: center;
  margin-right: 6px;
  display: inline-block;
}

.shortcut-item span {
  color: #666;
  flex: 1;
}

code {
  background: #f9f9f9;
  border: 1px solid #ddd;
  border-radius: 3px;
  padding: 2px 6px;
  font-family: monospace;
  font-size: 11px;
  color: #2c3e50;
  white-space: nowrap;
  margin-right: 6px;
  display: inline-block;
}

.close-btn {
  background: none;
  border: none;
  font-size: 28px;
  cursor: pointer;
  padding: 5px;
  color: #666;
  border-radius: 50%;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  line-height: 1;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #333;
}

@media (max-width: 768px) {
  .help-columns {
    grid-template-columns: 1fr;
  }

  .help-dialog {
    width: 95%;
    max-height: 90vh;
  }
}
</style>
