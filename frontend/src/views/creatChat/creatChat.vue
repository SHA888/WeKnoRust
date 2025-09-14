<template>
    <div v-show="cardList.length" class="dialogue-wrap">
        <div class="dialogue-answers">
            <div class="dialogue-title">
                <span>Knowledge-based Q&A</span>
            </div>
            <InputField @send-msg="sendMsg"></InputField>
        </div>
    </div>
    <EmptyKnowledge v-show="!cardList.length"></EmptyKnowledge>
</template>
<script setup lang="ts">
import { ref, onUnmounted, watch } from 'vue';
import InputField from '@/components/Input-field.vue';
import EmptyKnowledge from '@/components/empty-knowledge.vue';
import { getSessionsList, createSessions, generateSessionsTitle } from "@/api/chat/index";
import { useMenuStore } from '@/stores/menu';
import { useRoute, useRouter } from 'vue-router';
import useKnowledgeBase from '@/hooks/useKnowledgeBase';
import { getTestData } from '@/utils/request';

let { cardList } = useKnowledgeBase()
const router = useRouter();
const usemenuStore = useMenuStore();
const sendMsg = (value: string) => {
    createNewSession(value);
}

async function createNewSession(value: string) {
    // Read Knowledge Base ID from localStorage settings
    const settingsStr = localStorage.getItem("WeKnoRust_settings") ?? localStorage.getItem("WeKnora_settings");
    let knowledgeBaseId = "";
    
    if (settingsStr) {
        try {
            const settings = JSON.parse(settingsStr);
            if (settings.knowledgeBaseId) {
                knowledgeBaseId = settings.knowledgeBaseId;
                createSessions({ knowledge_base_id: knowledgeBaseId }).then(res => {
                    if (res.data && res.data.id) {
                        getTitle(res.data.id, value);
                    } else {
                        // Error handling
                        console.error("Failed to create session");
                    }
                }).catch(error => {
                    console.error("Error creating session:", error);
                });
                return;
            }
        } catch (e) {
            console.error("Failed to parse settings:", e);
        }
    }
    
    // If no Knowledge Base ID in settings, use test data
    const testData = getTestData();
    if (!testData || testData.knowledge_bases.length === 0) {
        console.error("Test data not initialized or contains no knowledge base");
        return;
    }

    // 使用第一个知识库ID
    knowledgeBaseId = testData.knowledge_bases[0].id;

    createSessions({ knowledge_base_id: knowledgeBaseId }).then(res => {
        if (res.data && res.data.id) {
            getTitle(res.data.id, value)
        } else {
            // Error handling
            console.error("Failed to create session");
        }
    }).catch(error => {
        console.error("Error creating session:", error);
    })
}

const getTitle = (session_id: string, value: string) => {
    let obj = { title: 'New Session', path: `chat/${session_id}`, id: session_id, isMore: false, isNoTitle: true }
    usemenuStore.updataMenuChildren(obj);
    usemenuStore.changeIsFirstSession(true);
    usemenuStore.changeFirstQuery(value);
    router.push(`/platform/chat/${session_id}`);
}

</script>
<style lang="less" scoped>
.dialogue-wrap {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    // position: relative;
}

.dialogue-answers {
    position: absolute;
    display: flex;
    flex-flow: column;
    align-items: center;

    :deep(.answers-input) {
        position: static;
        transform: translateX(0);
    }
}

.dialogue-title {
    display: flex;
    color: #000000;
    font-family: "PingFang SC";
    font-size: 28px;
    font-weight: 600;
    align-items: center;
    margin-bottom: 30px;

    .icon {
        display: flex;
        width: 32px;
        height: 32px;
        justify-content: center;
        align-items: center;
        border-radius: 6px;
        background: #FFF;
        box-shadow: 0 0 2px -1px #0000001f;
        margin-right: 12px;

        .logo_img {
            height: 24px;
            width: 24px;
        }
    }
}

@media (max-width: 1250px) and (min-width: 1045px) {
    .answers-input {
        transform: translateX(-329px);
    }

    :deep(.t-textarea__inner) {
        width: 654px !important;
    }
}

@media (max-width: 1045px) {
    .answers-input {
        transform: translateX(-250px);
    }

    :deep(.t-textarea__inner) {
        width: 500px !important;
    }
}
@media (max-width: 750px) {
    .answers-input {
        transform: translateX(-250px);
    }

    :deep(.t-textarea__inner) {
        width: 340px !important;
    }
}
@media (max-width: 600px) {
    .answers-input {
        transform: translateX(-250px);
    }

    :deep(.t-textarea__inner) {
        width: 300px !important;
    }
}

</style>
<style lang="less">
.del-menu-popup {
    z-index: 99 !important;

    .t-popup__content {
        width: 100px;
        height: 40px;
        line-height: 30px;
        padding-left: 14px;
        cursor: pointer;
        margin-top: 4px !important;

    }
}
</style>