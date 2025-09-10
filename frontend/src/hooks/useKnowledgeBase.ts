import { ref, reactive, onMounted } from "vue";
import { storeToRefs } from "pinia";
import { formatStringDate, kbFileTypeVerification } from "../utils/index";
import { MessagePlugin } from "tdesign-vue-next";
import {
  uploadKnowledgeBase,
  getKnowledgeBase,
  getKnowledgeDetails,
  delKnowledgeDetails,
  getKnowledgeDetailsCon,
} from "@/api/knowledge-base/index";
import { knowledgeStore } from "@/stores/knowledge";
const usemenuStore = knowledgeStore();
export default function () {
  const { cardList, total } = storeToRefs(usemenuStore);
  let moreIndex = ref(-1);
  const details = reactive({
    title: "",
    time: "",
    md: [],
    id: "",
    total: 0
  });
  const getKnowled = (query = { page: 1, page_size: 35 }) => {
    getKnowledgeBase(query)
      .then((result: any) => {
        let { data, total: totalResult } = result;
        let cardList_ = data.map((item) => {
            item["file_name"] = item.file_name.substring(
              0,
              item.file_name.lastIndexOf(".")
            );
            return {
              ...item,
              updated_at: formatStringDate(new Date(item.updated_at)),
              isMore: false,
              file_type: item.file_type.toLocaleUpperCase(),
            };
        });
        if (query.page == 1) {
          cardList.value = cardList_;
        } else {
          cardList.value.push(...cardList_);
        }
        total.value = totalResult;
      })
      .catch((err) => {});
  };
  const delKnowledge = (index: number, item) => {
    cardList.value[index].isMore = false;
    moreIndex.value = -1;
    delKnowledgeDetails(item.id)
      .then((result: any) => {
        if (result.success) {
          MessagePlugin.info("Knowledge deleted successfully!");
          getKnowled();
        } else {
          MessagePlugin.error("Failed to delete knowledge.");
        }
      })
      .catch((err) => {
        MessagePlugin.error("Failed to delete knowledge.");
      });
  };
  const openMore = (index: number) => {
    moreIndex.value = index;
  };
  const onVisibleChange = (visible: boolean) => {
    if (!visible) {
      moreIndex.value = -1;
    }
  };
  const requestMethod = (file: any, uploadInput) => {
    if (file instanceof File && uploadInput) {
      if (kbFileTypeVerification(file)) {
        return;
      }
      uploadKnowledgeBase({ file })
        .then((result: any) => {
          if (result.success) {
            MessagePlugin.info("Upload successful!");
            getKnowled();
          } else {
            // Improved error extraction
            let errorMessage = "Upload failed.";
            
            // Prefer message from error object
            if (result.error && result.error.message) {
              errorMessage = result.error.message;
            } else if (result.message) {
              errorMessage = result.message;
            }
            
            // Check error code for duplicate file
            if (result.code === 'duplicate_file' || (result.error && result.error.code === 'duplicate_file')) {
              errorMessage = "File already exists";
            }
            
            MessagePlugin.error(errorMessage);
          }
          uploadInput.value.value = "";
        })
        .catch((err: any) => {
          // Improved error handling in catch
          let errorMessage = "Upload failed.";
          
          if (err.code === 'duplicate_file') {
            errorMessage = "File already exists";
          } else if (err.error && err.error.message) {
            errorMessage = err.error.message;
          } else if (err.message) {
            errorMessage = err.message;
          }
          
          MessagePlugin.error(errorMessage);
          uploadInput.value.value = "";
        });
    } else {
      MessagePlugin.error("Invalid file type.");
    }
  };
  const getCardDetails = (item) => {
    Object.assign(details, {
      title: "",
      time: "",
      md: [],
      id: "",
    });
    getKnowledgeDetails(item.id)
      .then((result: any) => {
        if (result.success && result.data) {
          let { data } = result;
          Object.assign(details, {
            title: data.file_name,
            time: formatStringDate(new Date(data.updated_at)),
            id: data.id,
          });
        }
      })
      .catch((err) => {});
      getfDetails(item.id, 1);
  };
  const getfDetails = (id, page) => {
    getKnowledgeDetailsCon(id, page)
      .then((result: any) => {
        if (result.success && result.data) {
          let { data, total: totalResult } = result;
          if (page == 1) {
            details.md = data;
          } else {
            details.md.push(...data);
          }
          details.total = totalResult;
        }
      })
      .catch((err) => {});
  };
  return {
    cardList,
    moreIndex,
    getKnowled,
    details,
    delKnowledge,
    openMore,
    onVisibleChange,
    requestMethod,
    getCardDetails,
    total,
    getfDetails,
  };
}
