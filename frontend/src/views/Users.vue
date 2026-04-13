<script setup>
import { onMounted, ref } from 'vue'
import { 
  Users,
  UserPlus,
  Settings,
  Trash2,
  ShieldAlert,
  Mail,
  Shield
} from 'lucide-vue-next'
import { useUserStore } from '../stores/users'
import { useToastStore } from '../stores/toast'
import InteractiveImage from '../components/InteractiveImage.vue'

const userStore = useUserStore()
const toast = useToastStore()
const showAddModal = ref(false)
const isEditing = ref(false)
const editingUserId = ref(null)

const newUsername = ref('')
const newEmail = ref('')
const newPassword = ref('')
const newRoleId = ref('')
const newStatus = ref(true)

const isSubmitting = ref(false)

onMounted(async () => {
  await userStore.fetchRoles()
  await userStore.fetchUsers()
})

const handleAddUser = async () => {
  if (!newUsername.value || !newEmail.value || !newRoleId.value) return
  if (!isEditing.value && !newPassword.value) {
    toast.error('PASSWORD REQUIRED FOR NEW USERS')
    return
  }

  isSubmitting.value = true
  
  const payload = {
    username: newUsername.value,
    email: newEmail.value,
    role_id: parseInt(newRoleId.value, 10),
    status: Boolean(newStatus.value),
  }

  if (newPassword.value) {
    payload.password = newPassword.value
  }

  try {
    if (isEditing.value) {
      await userStore.updateUser(editingUserId.value, payload)
      toast.success(`USER [${payload.username}] UPDATED`)
    } else {
      await userStore.addUser(payload)
      toast.success(`NEW USER [${payload.username}] REGISTERED`)
    }
    showAddModal.value = false
    resetForm()
  } catch (err) {
    toast.error(`OPERATION FAILED: ${err.response?.data?.error || 'Unable to process request.'}`)
  } finally {
    isSubmitting.value = false
  }
}

const openEditModal = (user) => {
  isEditing.value = true
  editingUserId.value = user.id !== undefined ? user.id : user.ID
  newUsername.value = user.username || ''
  newEmail.value = user.email || ''
  newPassword.value = '' // don't load password
  
  // Handle role mapping since backend returns role name string instead of ID
  const matchedRole = userStore.roles.find(r => r.name === user.role)
  newRoleId.value = matchedRole ? (matchedRole.id !== undefined ? matchedRole.id : matchedRole.ID) : ''
  
  newStatus.value = user.status !== undefined ? user.status : true
  showAddModal.value = true
}

const handleDeleteUser = async (user) => {
  if (!confirm(`CONFIRM DELETION OF USER [${user.username}]? THIS ACTION IS IRREVERSIBLE.`)) return

  try {
    await userStore.deleteUser(user.id || user.ID)
    toast.success(`USER [${user.username}] TERMINATED`)
  } catch (err) {
    toast.error(`TERMINATION FAILED: ${err.response?.data?.error || 'Unable to process request.'}`)
  }
}

const resetForm = () => {
  isEditing.value = false
  editingUserId.value = null
  newUsername.value = ''
  newEmail.value = ''
  newPassword.value = ''
  newRoleId.value = ''
  newStatus.value = true
}

const getRoleName = (roleId) => {
  const role = userStore.roles.find(r => r.id === roleId || r.ID === roleId)
  return role ? role.name : 'Unknown Role'
}
</script>

<template>
  <div class="users-view">
    <div class="header-row">
      <h2 class="glow-cyan">USER MANAGEMENT GRID</h2>
      <div class="actions">
        <button @click="resetForm(); showAddModal = true" class="tron-btn">
          <UserPlus :size="18" />
          REGISTER USER
        </button>
      </div>
    </div>

    <!-- User Grid -->
    <div class="user-grid">
      <div 
        v-for="user in userStore.users" 
        :key="user.id || user.ID" 
        class="tron-card user-card"
        :class="{ 'active-user': user.status, 'inactive-user': !user.status }"
      >
        <div class="card-header">
           <div class="icon-wrap avatar-icon-wrap" :class="{'has-avatar': user.image_path && user.image_path.includes('/uploads/')}">
             <InteractiveImage v-if="user.image_path && user.image_path.includes('/uploads/')" :src="user.image_path" customClass="grid-avatar" alt="Avatar" />
             <Users v-else :size="24" :style="{ color: user.status ? 'var(--neon-cyan)' : 'var(--text-secondary)' }" />
           </div>
           <div class="user-info">
             <div class="title-row">
               <h3>{{ user.username }}</h3>
               <div class="action-buttons">
                 <button @click="openEditModal(user)" class="icon-btn edit-btn" title="EDIT USER">
                   <Settings :size="16" />
                 </button>
                 <button @click="handleDeleteUser(user)" class="icon-btn delete-btn" title="TERMINATE USER">
                   <Trash2 :size="16" />
                 </button>
               </div>
             </div>
             <a :href="'mailto:' + user.email" class="email-link font-data">
               <Mail :size="12" />
               {{ user.email }}
             </a>
           </div>
        </div>

        <div class="card-body">
           <div class="status-indicator">
              <span class="status-label">STATUS:</span>
              <span class="status-val" :style="{ color: user.status ? 'var(--neon-cyan)' : 'var(--neon-orange)' }">
                {{ user.status ? 'ACTIVE' : 'SUSPENDED' }}
              </span>
           </div>
           
           <div class="metrics-row">
              <div class="metric">
                <Shield :size="14" />
                <span>ROLE: {{ user.Role ? user.Role.name : getRoleName(user.role_id) }}</span>
              </div>
           </div>
        </div>
      </div>

      <div v-if="userStore.users.length === 0 && !userStore.loading" class="empty-state">
         <ShieldAlert :size="48" class="pulse" />
         <p>NO PERSONNEL DATA FOUND IN THE SYSTEM.</p>
         <button @click="resetForm(); showAddModal = true" class="tron-btn">REGISTER NEW PERSONNEL</button>
      </div>
    </div>

    <!-- Add User Modal -->
    <transition name="modal">
      <div v-if="showAddModal" class="modal-overlay">
        <div class="tron-card modal-container enhanced-modal">
          <div class="modal-header">
            <h3>{{ isEditing ? 'MODIFY PERSONNEL RECORD' : 'REGISTER NEW PERSONNEL' }}</h3>
          </div>
          <form @submit.prevent="handleAddUser" class="add-form">
            <div class="grid-inputs">
              <div class="input-group">
                <label>USERNAME</label>
                <input v-model="newUsername" type="text" placeholder="e.g. jdoe" required />
              </div>
              <div class="input-group">
                <label>EMAIL</label>
                <input v-model="newEmail" type="email" placeholder="jdoe@blackwater.sys" required />
              </div>
            </div>

            <div class="grid-inputs">
              <div class="input-group">
                <label>PASSWORD <span v-if="isEditing" class="text-secondary">(Leave blank to keep unchanged)</span></label>
                <input v-model="newPassword" type="password" placeholder="********" />
              </div>
              <div class="input-group">
                <label>ASSIGNED ROLE</label>
                <select v-model="newRoleId" required>
                  <option value="" disabled selected>Select Clearance Level</option>
                  <option v-for="role in userStore.roles" :key="role.id || role.ID" :value="role.id || role.ID">
                    {{ role.name }}
                  </option>
                </select>
              </div>
            </div>
            
            <div class="input-group checkbox-group">
              <label class="custom-checkbox">
                <input type="checkbox" v-model="newStatus" />
                <span class="checkmark"></span>
                ACCOUNT ACTIVE
              </label>
            </div>

            <div class="modal-footer">
              <button type="button" @click="showAddModal = false; resetForm()" class="tron-btn ghost">ABORT</button>
              <button type="submit" :disabled="isSubmitting" class="tron-btn">
                <span v-if="isSubmitting">{{ isEditing ? 'UPDATING...' : 'REGISTERING...' }}</span>
                <span v-else>{{ isEditing ? 'SAVE RECORD' : 'REGISTER PERSONNEL' }}</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>
  </div>
</template>

<style scoped>
.users-view {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.user-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 2rem;
}

@media (max-width: 768px) {
  .user-grid { gap: 1.5rem; }
}

.user-card {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  border-top: 3px solid transparent;
}

.active-user { border-top-color: var(--neon-cyan); }
.inactive-user { border-top-color: var(--neon-orange); }

.card-header {
  display: flex;
  align-items: center;
  gap: 1.2rem;
}

.icon-wrap {
  background: rgba(255, 255, 255, 0.03);
  padding: 0.8rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-wrap.has-avatar {
  padding: 0;
  width: 52px;
  height: 52px;
  border-radius: 50%;
  overflow: hidden;
  border: 1px solid var(--neon-cyan);
}

.grid-avatar {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-info {
  flex: 1;
}

.user-info h3 {
  font-size: 1.1rem;
  margin-bottom: 0.2rem;
  text-transform: uppercase;
}

.title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.icon-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  opacity: 0.6;
}

.icon-btn:hover {
  opacity: 1;
}

.edit-btn:hover {
  color: var(--neon-cyan);
  text-shadow: 0 0 8px var(--neon-cyan-glow);
}

.delete-btn:hover {
  color: var(--neon-orange);
  text-shadow: 0 0 8px var(--neon-orange-glow);
}

.email-link {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.email-link:hover { color: var(--neon-cyan); }

.card-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.status-indicator {
  display: flex;
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.02);
  justify-content: space-between;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.status-label { font-size: 0.8rem; color: var(--text-secondary); }
.status-val { font-weight: 700; letter-spacing: 2px; font-size: 0.9rem; }

.metrics-row {
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.metric { display: flex; align-items: center; gap: 0.5rem; }

.empty-state {
  grid-column: 1 / -1;
  padding: 5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  color: var(--text-secondary);
}

/* Modal Styling */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-container {
  width: 100%;
  max-width: 550px;
  padding: 2rem;
  margin: 1rem;
}

.add-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  margin-top: 1rem;
}

.input-group label {
  display: block;
  font-size: 0.75rem;
  margin-bottom: 0.5rem;
  color: var(--text-secondary);
  letter-spacing: 2px;
}

.input-group input:not([type="checkbox"]),
.input-group select {
  width: 100%;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(0, 242, 255, 0.2);
  padding: 0.8rem;
  color: var(--text-primary);
  font-family: var(--font-header);
  outline: none;
}

.input-group input:focus,
.input-group select:focus { 
  border-color: var(--neon-cyan); 
  box-shadow: 0 0 10px var(--neon-cyan-glow); 
}

/* Checkbox Style */
.checkbox-group {
  margin-top: 0.5rem;
}

.custom-checkbox {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  cursor: pointer;
  font-size: 0.8rem;
  color: var(--text-primary);
  position: relative;
  user-select: none;
}

.custom-checkbox input {
  position: absolute;
  opacity: 0;
  cursor: pointer;
  height: 0;
  width: 0;
}

.checkmark {
  height: 20px;
  width: 20px;
  background-color: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(0, 242, 255, 0.2);
  display: flex;
  align-items: center;
  justify-content: center;
}

.custom-checkbox:hover input ~ .checkmark {
  border-color: var(--neon-cyan);
}

.custom-checkbox input:checked ~ .checkmark {
  background-color: var(--neon-cyan);
  border-color: var(--neon-cyan);
  box-shadow: 0 0 10px var(--neon-cyan-glow);
}

.checkmark:after {
  content: "";
  position: absolute;
  display: none;
}

.custom-checkbox input:checked ~ .checkmark:after {
  display: block;
}

.custom-checkbox .checkmark:after {
  width: 5px;
  height: 10px;
  border: solid var(--bg-black);
  border-width: 0 2px 2px 0;
  transform: rotate(45deg);
}

.grid-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 600px) {
  .grid-inputs {
    grid-template-columns: 1fr;
    gap: 1rem;
  }
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1rem;
}
</style>
