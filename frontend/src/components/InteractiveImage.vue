<script setup>
import { ref } from 'vue'
import { Download, X, ZoomIn } from 'lucide-vue-next'
import { useToastStore } from '../stores/toast'

const props = defineProps({
  src: {
    type: String,
    required: true
  },
  alt: {
    type: String,
    default: 'Image'
  },
  customClass: {
    type: String,
    default: ''
  }
})

const isZoomed = ref(false)
const isDownloading = ref(false)
const toast = useToastStore()

const toggleZoom = () => {
  isZoomed.value = !isZoomed.value
}

const downloadImage = async () => {
  if (isDownloading.value) return
  isDownloading.value = true
  
  try {
    const response = await fetch(props.src)
    if (!response.ok) {
      throw new Error(`Server returned ${response.status} when accessing image.`)
    }
    const blob = await response.blob()
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    // Generate filename from URL
    const filename = props.src.split('/').pop() || 'downloaded_image.png'
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
  } catch (error) {
    console.error('Download failed:', error)
    // Likely a CORS issue or authorization error
    toast.error(`DOWNLOAD FAILED: Access to secure image asset rejected.`)
  } finally {
    isDownloading.value = false
  }
}
</script>

<template>
  <div class="interactive-image-container" :class="[customClass, { zoomed: isZoomed }]" @click.stop>
    <!-- Base Image Thumbnail -->
    <div class="thumbnail-wrapper" @click="toggleZoom">
      <img :src="src" :alt="alt" class="thumbnail-target" />
      <div class="hover-overlay">
        <ZoomIn :size="24" class="zoom-icon" />
      </div>
    </div>

    <!-- Zoomed Analytics Modal -->
    <Teleport to="body">
      <transition name="fade">
        <div v-if="isZoomed" class="image-modal-overlay" @click="toggleZoom">
          <div class="modal-controls">
            <button class="tron-btn control-btn" title="DOWNLOAD ASSET" @click.stop="downloadImage" :disabled="isDownloading">
              <Download :size="20" />
              <span>{{ isDownloading ? 'EXTRACTING...' : 'SAVE DATA' }}</span>
            </button>
            <button class="tron-btn ghost control-btn close-btn" title="CLOSE" @click.stop="toggleZoom">
              <X :size="24" />
            </button>
          </div>
          
          <div class="zoomed-asset-container" @click.stop>
             <img :src="src" :alt="alt" class="zoomed-target" />
             <div class="scanline"></div>
          </div>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<style scoped>
.interactive-image-container {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: inherit;
}

.thumbnail-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: inherit;
  overflow: hidden;
  cursor: zoom-in;
}

.thumbnail-target {
  width: 100%;
  height: 100%;
  object-fit: inherit;
  transition: transform 0.3s ease;
}

.hover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.zoom-icon {
  color: var(--neon-cyan);
  filter: drop-shadow(0 0 5px var(--neon-cyan-glow));
  transform: scale(0.8);
  transition: transform 0.2s ease;
}

.thumbnail-wrapper:hover .hover-overlay {
  opacity: 1;
}

.thumbnail-wrapper:hover .zoom-icon {
  transform: scale(1);
}

.thumbnail-wrapper:hover .thumbnail-target {
  transform: scale(1.05);
}

/* Modal Overlay */
.image-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(5, 7, 10, 0.95);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(10px);
}

.modal-controls {
  position: absolute;
  top: 2rem;
  right: 2rem;
  display: flex;
  gap: 1rem;
  z-index: 10000;
}

.control-btn {
  padding: 0.5rem 1rem !important;
}

.close-btn {
  padding: 0.5rem !important;
  color: var(--neon-orange);
  border-color: var(--neon-orange);
}

.close-btn:hover {
  background: var(--neon-orange);
  color: #fff;
  box-shadow: 0 0 10px var(--neon-orange-glow);
}

.zoomed-asset-container {
  position: relative;
  max-width: 90vw;
  max-height: 80vh;
  border: 1px solid rgba(0, 242, 255, 0.3);
  box-shadow: 0 0 30px rgba(0, 242, 255, 0.1);
  overflow: hidden;
  user-select: none;
}

.zoomed-target {
  max-width: 100%;
  max-height: 80vh;
  display: block;
  object-fit: contain;
}

/* Sci-fi Scanline effect over full screen image */
.scanline {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: var(--neon-cyan);
  opacity: 0.3;
  box-shadow: 0 0 10px var(--neon-cyan-glow);
  animation: scan 3s linear infinite;
  pointer-events: none;
}

@keyframes scan {
  0% { transform: translateY(0); }
  100% { transform: translateY(80vh); opacity: 0; }
}
</style>
