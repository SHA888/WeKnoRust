<template>
    <div class="settings-container">
        <div class="settings-header">
            <h2>System Configuration</h2>
        </div>
        <div class="settings-form">
            <t-form ref="form" :data="formData" :rules="rules" @submit="onSubmit">
                <t-form-item label="API Endpoint" name="endpoint">
                    <t-input v-model="formData.endpoint" placeholder="Enter API endpoint, e.g. http://localhost" />
                </t-form-item>
                <t-form-item label="API Key" name="apiKey">
                    <t-input v-model="formData.apiKey" placeholder="Enter API Key" />
                </t-form-item>
                <t-form-item label="Knowledge Base ID" name="knowledgeBaseId">
                    <t-input v-model="formData.knowledgeBaseId" placeholder="Enter Knowledge Base ID" />
                </t-form-item>
                <t-form-item>
                    <t-space>
                        <t-button theme="primary" type="submit">Save</t-button>
                        <t-button theme="default" @click="resetForm">Reset</t-button>
                    </t-space>
                </t-form-item>
            </t-form>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { useSettingsStore } from '@/stores/settings';

const settingsStore = useSettingsStore();
const form = ref(null);

const formData = reactive({
    endpoint: '',
    apiKey: '',
    knowledgeBaseId: ''
});

const rules = {
    endpoint: [{ required: true, message: 'Please enter API endpoint', trigger: 'blur' }],
    apiKey: [{ required: true, message: 'Please enter API Key', trigger: 'blur' }],
    knowledgeBaseId: [{ required: true, message: 'Please enter Knowledge Base ID', trigger: 'blur' }]
};

onMounted(() => {
    // Initialize form data
    const settings = settingsStore.getSettings();
    formData.endpoint = settings.endpoint;
    formData.apiKey = settings.apiKey;
    formData.knowledgeBaseId = settings.knowledgeBaseId;
});

const onSubmit = ({ validateResult }) => {
    if (validateResult === true) {
        settingsStore.saveSettings({
            endpoint: formData.endpoint,
            apiKey: formData.apiKey,
            knowledgeBaseId: formData.knowledgeBaseId
        });
        MessagePlugin.success('Settings saved successfully');
    }
};

const resetForm = () => {
    const settings = settingsStore.getSettings();
    formData.endpoint = settings.endpoint;
    formData.apiKey = settings.apiKey;
    formData.knowledgeBaseId = settings.knowledgeBaseId;
};
</script>

<style lang="less" scoped>
.settings-container {
    padding: 20px;
    background-color: #fff;
    border-radius: 8px;
    margin: 20px;
    min-height: 80vh;

    .settings-header {
        margin-bottom: 20px;
        border-bottom: 1px solid #f0f0f0;
        padding-bottom: 16px;

        h2 {
            font-size: 20px;
            font-weight: 600;
            color: #000000;
            margin: 0;
        }
    }

    .settings-form {
        max-width: 600px;
    }
}
</style> 