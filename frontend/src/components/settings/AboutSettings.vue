<script setup lang="ts">
import { Icon } from '@iconify/vue';
</script>

<template>
  <div>
    <h2>MrRss</h2>
    <h4>{{ $t('Settings.AboutSettings.description') }}</h4>
  </div>
  <img src="../../assets/images/appicon.png" class="logo" :alt="$t('Settings.AboutSettings.appLogo')" />
  <div class="version">
    <button class="btn" :title="$t('Settings.AboutSettings.checkUpdate')" disabled >
      <Icon icon="material-symbols:deployed-code-update-outline" />
    </button>
    <p>
      {{ $t('Settings.AboutSettings.version', {version: '0.2.0'}) }}
    </p>
  </div>
  <div class="copyright">
    <p>© 2024 <a href="https://github.com/WCY-dt/MrRSS" rel="help" target="_blank">MrRss</a>. {{
      $t('Settings.AboutSettings.copyright') }}
    </p>
    <p>{{ $t('Settings.AboutSettings.designAndDev') }}<a href="https://ch3nyang.top/about/" rel="help"
        target="_blank">Ch3nyang</a>.
    </p>
  </div>
</template>

<style lang="scss" scoped>
@import '../../styles/settings/AboutSettings.scss';
</style>