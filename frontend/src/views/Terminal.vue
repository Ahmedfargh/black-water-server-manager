<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Terminal as TerminalIcon, Power, Trash2, Clock } from 'lucide-vue-next'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const authStore = useAuthStore()

const output = ref([{ type: 'system', text: t('terminal.init_msg') || 'BLACKWATER TERMINAL SYSTEM v1.0.0 INITIALIZED.' }])
const commandInput = ref('')
const terminalRef = ref(null)
const inputRef = ref(null)
const wsStatus = ref('DISCONNECTED')
const history = ref([])
const historyIndex = ref(-1)

let ws = null

const scrollToBottom = () => {
  nextTick(() => {
    if (terminalRef.value) {
      terminalRef.value.scrollTop = terminalRef.value.scrollHeight
    }
  })
}

const connectWebSocket = () => {
  if (ws) ws.close()
  
  wsStatus.value = 'CONNECTING'
  output.value.push({ type: 'system', text: t('terminal.handshake') || 'INITIATING HANDSHAKE WITH SERVER...' })
  
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${protocol}//${window.location.host}/ws/terminal?token=${authStore.token}`
  
  try {
    ws = new WebSocket(wsUrl)
    
    ws.onopen = () => {
      wsStatus.value = 'CONNECTED'
      output.value.push({ type: 'success', text: t('terminal.access_granted') || 'SECURE CONNECTION ESTABLISHED. ACCESS GRANTED.' })
      scrollToBottom()
      focusInput()
    }
    
    ws.onmessage = (event) => {
      output.value.push({ type: 'response', text: event.data })
      scrollToBottom()
    }
    
    ws.onclose = () => {
      wsStatus.value = 'DISCONNECTED'
      output.value.push({ type: 'error', text: t('terminal.conn_terminated') || 'CONNECTION TERMINATED BY REMOTE HOST.' })
      scrollToBottom()
    }

    ws.onerror = (error) => {
      wsStatus.value = 'ERROR'
      output.value.push({ type: 'error', text: t('terminal.socket_error') || 'SOCKET ERROR DETECTED.' })
      console.error("WebSocket Error: ", error)
      scrollToBottom()
    }
  } catch (err) {
    output.value.push({ type: 'error', text: `${t('terminal.conn_failed') || 'FAILED TO CONNECT'}: ${err.message}` })
    wsStatus.value = 'ERROR'
  }
}

const disconnectWebSocket = () => {
  if (ws) {
    ws.close()
  }
}

onMounted(() => {
  connectWebSocket()
})

onUnmounted(() => {
  disconnectWebSocket()
})

const sendCommand = () => {
  const cmd = commandInput.value.trim()
  if (!cmd) return
  
  // Echo command to output
  output.value.push({ type: 'command', text: `${authStore.user?.username || 'admin'}@blackwater:~$ ${cmd}` })
  
  if (cmd.toLowerCase() === 'clear') {
    output.value = []
    commandInput.value = ''
    return
  }

  // Handle command history
  history.value.push(cmd)
  historyIndex.value = history.value.length

  if (ws && ws.readyState === WebSocket.OPEN) {
    const payload = JSON.stringify({ command: cmd })
    ws.send(payload)
  } else {
    output.value.push({ type: 'error', text: t('terminal.not_connected') || 'NOT CONNECTED TO SERVER. COMMAND IGNORED.' })
  }
  
  commandInput.value = ''
  scrollToBottom()
}

const handleKeyDown = (e) => {
  if (e.key === 'ArrowUp') {
    e.preventDefault()
    if (historyIndex.value > 0) {
      historyIndex.value--
      commandInput.value = history.value[historyIndex.value]
    }
  } else if (e.key === 'ArrowDown') {
    e.preventDefault()
    if (historyIndex.value < history.value.length - 1) {
      historyIndex.value++
      commandInput.value = history.value[historyIndex.value]
    } else {
      historyIndex.value = history.value.length
      commandInput.value = ''
    }
  }
}

const clearTerminal = () => {
  output.value = []
}

const focusInput = () => {
  if (inputRef.value) {
    inputRef.value.focus()
  }
}

// Localize status badge text
const getStatusText = (status) => {
  if (status === 'CONNECTED') return t('terminal.connected') || 'CONNECTED'
  if (status === 'DISCONNECTED') return t('terminal.disconnected') || 'DISCONNECTED'
  if (status === 'CONNECTING') return t('terminal.connecting') || 'CONNECTING...'
  if (status === 'ERROR') return t('terminal.error') || 'ERROR'
  return status
}
</script>

<template>
  <div class="terminal-view">
    <div class="terminal-header">
      <div class="header-left">
        <TerminalIcon class="glow-cyan" :size="24" />
        <h2 class="pulse-text font-header">{{ $t('terminal.sys_terminal') }}</h2>
        <span class="status-badge" :class="wsStatus.replace('...', '').toLowerCase()">
          {{ getStatusText(wsStatus) }}
        </span>
      </div>
      <div class="header-actions">
        <button @click="clearTerminal" class="action-btn outline-cyan" :title="$t('terminal.clear_log')">
          <Trash2 :size="16" />
          <span>{{ $t('terminal.clear') }}</span>
        </button>
        <button v-if="wsStatus !== 'CONNECTED'" @click="connectWebSocket" class="action-btn fill-cyan" :title="$t('terminal.reconnect')">
          <Power :size="16" />
          <span>{{ $t('terminal.connect') }}</span>
        </button>
        <button v-else @click="disconnectWebSocket" class="action-btn border-orange" :title="$t('terminal.disconnect_btn')">
          <Power :size="16" />
          <span>{{ $t('terminal.disconnect') }}</span>
        </button>
      </div>
    </div>

    <div class="terminal-container tron-card" @click="focusInput" dir="ltr">
      <div class="terminal-output" ref="terminalRef">
        <div 
          v-for="(line, index) in output" 
          :key="index"
          class="log-line font-data"
          :class="`line-${line.type}`"
        >
          <pre>{{ line.text }}</pre>
        </div>
      </div>
      
      <div class="terminal-input-row" v-if="wsStatus === 'CONNECTED'">
        <span class="prompt font-data text-secondary">{{ authStore.user?.username || 'admin' }}@blackwater:~$</span>
        <input 
          ref="inputRef"
          v-model="commandInput"
          type="text"
          class="command-input font-data"
          autocomplete="off"
          spellcheck="false"
          @keyup.enter="sendCommand"
          @keydown="handleKeyDown"
        />
      </div>
      <div class="terminal-input-row disabled" v-else>
         <span class="prompt font-data neon-orange">{{ $t('terminal.offline_msg') }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.terminal-view {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  height: 100%;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 1rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

h2 {
  margin: 0;
  color: var(--text-primary);
  letter-spacing: 2px;
}

.status-badge {
  font-size: 0.75rem;
  padding: 0.3rem 0.8rem;
  border-radius: 4px;
  font-weight: 600;
  letter-spacing: 1px;
}

.status-badge.connected {
  background: rgba(0, 242, 255, 0.1);
  color: var(--neon-cyan);
  border: 1px solid var(--neon-cyan-glow);
  box-shadow: 0 0 10px rgba(0, 242, 255, 0.2);
}

.status-badge.disconnected, .status-badge.error {
  background: rgba(255, 60, 0, 0.1);
  color: var(--neon-orange);
  border: 1px solid var(--neon-orange-glow);
}

.status-badge.connecting {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  border: 1px solid rgba(255, 255, 255, 0.2);
}

.header-actions {
  display: flex;
  gap: 1rem;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 1rem;
  border-radius: 4px;
  font-family: inherit;
  font-weight: 600;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.3s ease;
  letter-spacing: 1px;
}

.outline-cyan {
  background: transparent;
  border: 1px solid rgba(0, 242, 255, 0.3);
  color: var(--neon-cyan);
}

.outline-cyan:hover {
  background: rgba(0, 242, 255, 0.1);
  border-color: var(--neon-cyan);
}

.fill-cyan {
  background: rgba(0, 242, 255, 0.15);
  border: 1px solid var(--neon-cyan);
  color: #fff;
  text-shadow: 0 0 5px var(--neon-cyan-glow);
  box-shadow: inset 0 0 10px rgba(0, 242, 255, 0.2);
}

.fill-cyan:hover {
  background: rgba(0, 242, 255, 0.3);
  box-shadow: 0 0 15px var(--neon-cyan-glow);
}

.border-orange {
  background: transparent;
  border: 1px solid var(--neon-orange);
  color: var(--neon-orange);
}

.border-orange:hover {
  background: rgba(255, 60, 0, 0.1);
  box-shadow: 0 0 10px var(--neon-orange-glow);
}

.terminal-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #050505;
  border: 1px solid rgba(0, 242, 255, 0.3);
  box-shadow: inset 0 0 30px rgba(0, 0, 0, 0.8), 0 0 15px rgba(0, 242, 255, 0.05);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  min-height: 500px;
  text-align: left;
}

.terminal-container::before {
  content: "";
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: linear-gradient(rgba(18, 16, 16, 0) 50%, rgba(0, 0, 0, 0.25) 50%), linear-gradient(90deg, rgba(255, 0, 0, 0.06), rgba(0, 255, 0, 0.02), rgba(0, 0, 255, 0.06));
  background-size: 100% 2px, 3px 100%;
  pointer-events: none;
  z-index: 10;
  opacity: 0.15;
}

.terminal-output {
  flex: 1;
  padding: 1.5rem;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.log-line {
  word-break: break-all;
  line-height: 1.4;
  margin: 0;
}

.log-line pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
}

.line-system {
  color: var(--text-secondary);
}

.line-success {
  color: #00ff00;
  text-shadow: 0 0 5px rgba(0, 255, 0, 0.3);
}

.line-error {
  color: var(--neon-orange);
}

.line-command {
  color: var(--neon-cyan);
  margin-top: 0.5rem;
}

.line-response {
  color: #a0aab5;
}

.terminal-input-row {
  display: flex;
  padding: 1rem 1.5rem;
  background: rgba(0, 242, 255, 0.03);
  border-top: 1px solid rgba(0, 242, 255, 0.1);
  align-items: center;
  gap: 0.8rem;
}

.terminal-input-row.disabled {
  background: rgba(255, 60, 0, 0.03);
  border-top-color: rgba(255, 60, 0, 0.1);
  justify-content: center;
}

.prompt {
  white-space: nowrap;
}

.command-input {
  flex: 1;
  background: transparent;
  border: none;
  color: #fff;
  font-size: 1rem;
  outline: none;
  caret-color: var(--neon-cyan);
}

.command-input:focus {
  outline: none;
}
</style>
