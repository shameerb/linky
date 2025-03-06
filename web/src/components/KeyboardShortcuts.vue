<template>
  <div class="keyboard-shortcuts-help" v-if="show">
    <div class="help-content">
      <div class="help-header">
        <h3>Keyboard Shortcuts</h3>
        <button class="close-btn" @click="close" title="Close">×</button>
      </div>
      <div class="shortcuts-list">
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
          <div class="shortcut-item">
            <kbd>a</kbd>
            <span>Select all visible links</span>
          </div>
          <div class="shortcut-item">
            <kbd>A</kbd>
            <span>Clear selection</span>
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
        <div class="shortcut-group">
          <h4>Other</h4>
          <div class="shortcut-item">
            <kbd>?</kbd>
            <span>Toggle help</span>
          </div>
          <div class="shortcut-item">
            <kbd>Ctrl</kbd> + <kbd>F</kbd>
            <span>Focus search</span>
          </div>
          <div class="shortcut-item">
            <kbd>Ctrl</kbd> + <kbd>O</kbd>
            <span>Focus file selector</span>
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
  methods: {
    close() {
      this.$emit('close')
    }
  }
}
</script>

<style scoped>
.keyboard-shortcuts-help {
  position: fixed;
  bottom: 20px;
  right: 20px;
  background: rgba(255, 255, 255, 0.85); /* Made more translucent */
  backdrop-filter: blur(8px); /* Increased blur effect */
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.15);
  z-index: 1000;
  width: 320px;
}

.help-content {
  padding: 16px;
}

.help-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.help-header h3 {
  margin: 0;
  color: #333;
}

.shortcut-group {
  margin-bottom: 16px;
}

.shortcut-group h4 {
  margin: 0 0 8px 0;
  color: #666;
  font-size: 0.9em;
}

.shortcut-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 0.9em;
}

kbd {
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 3px;
  padding: 2px 6px;
  box-shadow: 0 1px 1px rgba(0,0,0,0.2);
  font-family: monospace;
  font-size: 0.9em;
  min-width: 20px;
  text-align: center;
  margin-right: 8px;
}

.shortcut-item span {
  color: #666;
}

.close-btn {
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

.close-btn:hover {
  background: #f0f0f0;
  color: #333;
}
</style> 