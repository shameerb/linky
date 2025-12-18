<template>
  <div class="auth-container">
    <div class="auth-card">
      <h1>Linky</h1>
      <h2>{{ isSignup ? 'Create Account' : 'Welcome Back' }}</h2>

      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label for="email">Email</label>
          <input
            type="email"
            id="email"
            v-model="email"
            placeholder="your@email.com"
            required
            autofocus
          />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input
            type="password"
            id="password"
            v-model="password"
            placeholder="Enter your password"
            required
            :minlength="isSignup ? 6 : 1"
          />
        </div>

        <div v-if="error" class="error-message">
          {{ error }}
        </div>

        <button type="submit" class="btn-primary" :disabled="loading">
          {{ loading ? 'Loading...' : (isSignup ? 'Sign Up' : 'Log In') }}
        </button>
      </form>

      <div class="auth-footer">
        <p v-if="!isSignup">
          Don't have an account?
          <a href="#" @click.prevent="toggleMode">Sign up</a>
        </p>
        <p v-else>
          Already have an account?
          <a href="#" @click.prevent="toggleMode">Log in</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Login',
  data() {
    return {
      email: '',
      password: '',
      isSignup: false,
      loading: false,
      error: ''
    }
  },
  methods: {
    async handleSubmit() {
      this.error = ''
      this.loading = true

      try {
        const endpoint = this.isSignup ? '/api/auth/signup' : '/api/auth/login'
        const response = await fetch(endpoint, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            email: this.email,
            password: this.password
          })
        })

        if (!response.ok) {
          const errorText = await response.text()
          throw new Error(errorText || 'Authentication failed')
        }

        const data = await response.json()

        // Store token and user info
        localStorage.setItem('token', data.token)
        localStorage.setItem('user', JSON.stringify(data.user))

        // Emit login event to parent
        this.$emit('login', data)
      } catch (err) {
        this.error = err.message
      } finally {
        this.loading = false
      }
    },
    toggleMode() {
      this.isSignup = !this.isSignup
      this.error = ''
      this.password = ''
    }
  }
}
</script>

<style scoped>
.auth-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.auth-card {
  background: white;
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.1);
  width: 100%;
  max-width: 400px;
}

.auth-card h1 {
  margin: 0 0 8px 0;
  font-size: 32px;
  font-weight: 700;
  color: #667eea;
  text-align: center;
}

.auth-card h2 {
  margin: 0 0 32px 0;
  font-size: 18px;
  font-weight: 400;
  color: #666;
  text-align: center;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #333;
}

.form-group input {
  width: 100%;
  padding: 12px 14px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  box-sizing: border-box;
  transition: all 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.error-message {
  padding: 12px;
  background-color: #fee2e2;
  border: 1px solid #fecaca;
  border-radius: 6px;
  color: #dc2626;
  font-size: 14px;
  margin-bottom: 20px;
}

.btn-primary {
  width: 100%;
  padding: 12px;
  background: #667eea;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary:hover:not(:disabled) {
  background: #5568d3;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.auth-footer {
  margin-top: 24px;
  text-align: center;
}

.auth-footer p {
  margin: 0;
  font-size: 14px;
  color: #666;
}

.auth-footer a {
  color: #667eea;
  text-decoration: none;
  font-weight: 500;
}

.auth-footer a:hover {
  text-decoration: underline;
}
</style>
