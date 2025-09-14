# Frequently Asked Questions (FAQ)

## 1. How do I view logs?
```bash
# View main service logs
docker exec -it WeKnoRust-app tail -f /var/log/WeKnoRust.log

# View doc-reader module logs
docker exec -it WeKnoRust-docreader tail -f /var/log/docreader.log
```

## 2. How do I start and stop the services?
```bash
# Start services
./scripts/start_all.sh

# Stop services
./scripts/start_all.sh --stop

# Wipe the database
./scripts/start_all.sh --stop && make clean-db
```

## 3. After starting services, uploads fail or the system doesn’t work as expected

This usually happens when the Embedding model and/or the Chat LLM are not configured correctly. Diagnose with the following steps.

1) Check that model settings are complete in `.env`. If using ollama for local models, ensure the local ollama service is running and the following environment variables are set correctly in `.env`:
```bash
# LLM Model
INIT_LLM_MODEL_NAME=your_llm_model
# Embedding Model
INIT_EMBEDDING_MODEL_NAME=your_embedding_model
# Embedding vector dimension
INIT_EMBEDDING_MODEL_DIMENSION=your_embedding_model_dimension
# Embedding model ID (usually a string)
INIT_EMBEDDING_MODEL_ID=your_embedding_model_id
```

If you are accessing models via a remote API, you must also provide the corresponding `BASE_URL` and `API_KEY`:
```bash
# LLM base URL
INIT_LLM_MODEL_BASE_URL=your_llm_model_base_url
# LLM API key (set if authentication is required)
INIT_LLM_MODEL_API_KEY=your_llm_model_api_key
# Embedding base URL
INIT_EMBEDDING_MODEL_BASE_URL=your_embedding_model_base_url
# Embedding API key (set if authentication is required)
INIT_EMBEDDING_MODEL_API_KEY=your_embedding_model_api_key
```

When reranking is required, configure the Rerank model as follows:
```bash
# Rerank model name
INIT_RERANK_MODEL_NAME=your_rerank_model_name
# Rerank base URL
INIT_RERANK_MODEL_BASE_URL=your_rerank_model_base_url
# Rerank API key (set if authentication is required)
INIT_RERANK_MODEL_API_KEY=your_rerank_model_api_key
```

2) Check the main service logs and see if there are any `ERROR` entries.

## 4. How do I enable multimodal features?
1) Ensure the following `.env` values are set:
```bash
# VLM_MODEL_NAME: Vision-Language model name
VLM_MODEL_NAME=your_vlm_model_name

# VLM_MODEL_BASE_URL: Vision-Language model base URL
VLM_MODEL_BASE_URL=your_vlm_model_base_url

# VLM_MODEL_API_KEY: Vision-Language model API key
VLM_MODEL_API_KEY=your_vlm_model_api_key
```
Note: Currently, multimodal LLMs are supported via remote API only, so `VLM_MODEL_BASE_URL` and `VLM_MODEL_API_KEY` must be provided.

2) Parsed files must be uploaded to COS (or your configured object storage). Ensure the COS-related variables in `.env` are set correctly:
```bash
# Tencent COS access key ID
COS_SECRET_ID=your_cos_secret_id

# Tencent COS secret key
COS_SECRET_KEY=your_cos_secret_key

# Tencent COS region, e.g., ap-guangzhou
COS_REGION=your_cos_region

# Tencent COS bucket name
COS_BUCKET_NAME=your_cos_bucket_name

# Tencent COS app ID
COS_APP_ID=your_cos_app_id

# Path prefix in COS for storing files
COS_PATH_PREFIX=your_cos_path_prefix
```
Important: Set files in COS to public read, otherwise the doc-reader module cannot access and parse them.

3) Check doc-reader module logs and verify that OCR and captioning are processed and logged as expected.

## P.S.
If the above steps don’t resolve your problem, please open an issue describing the problem and provide relevant logs to help us diagnose.
