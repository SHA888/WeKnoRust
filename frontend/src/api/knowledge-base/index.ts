import { get, post, put, del, postUpload, getDown, getTestData } from "../../utils/request";
import { loadTestData } from "../test-data";

// Get Knowledge Base ID (prefer from settings)
async function getKnowledgeBaseID() {
  // Read Knowledge Base ID from localStorage settings
  const settingsStr = localStorage.getItem("WeKnoRust_settings") ?? localStorage.getItem("WeKnora_settings");
  let knowledgeBaseId = "";
  
  if (settingsStr) {
    try {
      const settings = JSON.parse(settingsStr);
      if (settings.knowledgeBaseId) {
        return settings.knowledgeBaseId;
      }
    } catch (e) {
      console.error("Failed to parse settings:", e);
    }
  }
  
  // If no Knowledge Base ID in settings, use test data
  await loadTestData();
  
  const testData = getTestData();
  if (!testData || testData.knowledge_bases.length === 0) {
    console.error("Test data not initialized or contains no knowledge base");
    throw new Error("Test data not initialized or contains no knowledge base");
  }
  return testData.knowledge_bases[0].id;
}

export async function uploadKnowledgeBase(data = {}) {
  const kbId = await getKnowledgeBaseID();
  return postUpload(`/api/v1/knowledge-bases/${kbId}/knowledge/file`, data);
}

export async function getKnowledgeBase({page, page_size}) {
  const kbId = await getKnowledgeBaseID();
  return get(
    `/api/v1/knowledge-bases/${kbId}/knowledge?page=${page}&page_size=${page_size}`
  );
}

export function getKnowledgeDetails(id: any) {
  return get(`/api/v1/knowledge/${id}`);
}

export function delKnowledgeDetails(id: any) {
  return del(`/api/v1/knowledge/${id}`);
}

export function downKnowledgeDetails(id: any) {
  return getDown(`/api/v1/knowledge/${id}/download`);
}

export function batchQueryKnowledge(ids: any) {
  return get(`/api/v1/knowledge/batch?${ids}`);
}

export function getKnowledgeDetailsCon(id: any, page) {
  return get(`/api/v1/chunks/${id}?page=${page}&page_size=25`);
}