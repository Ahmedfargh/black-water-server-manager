<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps({
  size: {
    type: String,
    default: 'md', // 'sm', 'md', 'lg'
    validator: (val) => ['sm', 'md', 'lg'].includes(val)
  },
  showText: {
    type: Boolean,
    default: true
  },
  badgeText: {
    type: String,
    default: 'LINUX // v1.24'
  }
})

const { t } = useI18n()

const iconSizes = {
  sm: 32,
  md: 44,
  lg: 84
}

const currentIconSize = computed(() => iconSizes[props.size] || 44)
</script>

<template>
  <div class="brand-logo" :class="[`size-${size}`]">
    <!-- Custom Frontier Outlaw Star & CLI Logo Emblem -->
    <div class="logo-mark" :style="{ width: `${currentIconSize}px`, height: `${currentIconSize}px` }">
      <img 
        src="/blackwater_logo.jpg" 
        alt="Blackwater Logo" 
        class="logo-img"
      />
    </div>

    <!-- Brand Typography & Subtitle -->
    <div v-if="showText" class="brand-text-block">
      <span class="brand-name">{{ $t('app.logo_text') }}</span>
      <span v-if="badgeText" class="brand-badge font-data">{{ badgeText }}</span>
    </div>
  </div>
</template>

<style scoped>
.brand-logo {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  user-select: none;
}

.logo-mark {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.6), 0 0 12px rgba(220, 38, 38, 0.25);
  border: 1px solid rgba(220, 38, 38, 0.4);
  transition: transform 0.25s ease, box-shadow 0.25s ease;
}

.brand-logo:hover .logo-mark {
  transform: scale(1.06) rotate(3deg);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.7), 0 0 18px rgba(220, 38, 38, 0.45);
}

.logo-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.brand-text-block {
  display: flex;
  flex-direction: column;
  line-height: 1.15;
}

.brand-name {
  font-family: var(--font-header);
  font-weight: 800;
  letter-spacing: 2px;
  color: #ffffff;
  text-transform: uppercase;
}

[dir="rtl"] .brand-name {
  font-family: var(--font-arabic);
  letter-spacing: 0px !important;
}

.brand-badge {
  font-size: 0.65rem;
  letter-spacing: 0.8px;
  color: var(--rdr-amber);
  font-weight: 600;
  opacity: 0.95;
  margin-top: 1px;
}

/* Sizing variations */
.size-sm .brand-name {
  font-size: 0.95rem;
  letter-spacing: 1.5px;
}
.size-sm .brand-badge {
  font-size: 0.6rem;
}

.size-md .brand-name {
  font-size: 1.15rem;
  letter-spacing: 2px;
}
.size-md .brand-badge {
  font-size: 0.65rem;
}

.size-lg {
  flex-direction: column;
  gap: 0.9rem;
  text-align: center;
}
.size-lg .brand-name {
  font-size: 1.8rem;
  letter-spacing: 3px;
}
.size-lg .brand-badge {
  font-size: 0.8rem;
  letter-spacing: 1.5px;
}
</style>
