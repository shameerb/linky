import { createApp } from 'vue'
import App from './App.vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faTrashAlt, faCopy, faPlus, faCog } from '@fortawesome/free-solid-svg-icons'

library.add(faTrashAlt, faCopy, faPlus, faCog)

const app = createApp(App)
app.component('font-awesome-icon', FontAwesomeIcon)
app.mount('#app')
