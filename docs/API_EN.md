# WeKnoRust API Documentation (EN)

## Table of Contents

- [Overview](#overview)
- [Basics](#basics)
- [Authentication](#authentication)
- [Error Handling](#error-handling)
- [API Categories](#api-categories)
- [Detailed APIs](#detailed-apis)
  - [Tenant Management](#tenant-management)
  - [Knowledge Base Management](#knowledge-base-management)
  - [Knowledge Management](#knowledge-management)
  - [Model Management](#model-management)
  - [Chunk Management](#chunk-management)
  - [Session Management](#session-management)
  - [Chat](#chat)
  - [Message Management](#message-management)
  - [Evaluation](#evaluation)

## Overview

WeKnoRust exposes a set of RESTful APIs for creating and managing knowledge bases, ingesting and retrieving knowledge, and performing knowledge-grounded Q&A. This document describes how to use these APIs with examples.

## Basics

- Base URL: `/api/v1`
- Response format: JSON
- Authentication: API Key via HTTP header

## Authentication

All requests must include `X-API-Key` in the request headers:

```
X-API-Key: your_api_key
```

For observability and traceability, we recommend adding a unique `X-Request-ID` header per request:

```
X-Request-ID: unique_request_id
```

### How to obtain an API Key

- When creating a new tenant via `POST /api/v1/tenants`, the response contains a newly generated API key. Keep your API key safe. It grants full access to your tenant's APIs.

## Error Handling

WeKnoRust uses standard HTTP status codes and a unified error response body:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": "Optional extra details"
  }
}
```

## API Categories

1. Tenant management: create and manage tenants
2. Knowledge base management: create, query, and manage knowledge bases
3. Knowledge management: upload, retrieve, and manage knowledge items
4. Model management: configure and manage AI models
5. Chunk management: manage chunked content
6. Session management: create and manage chat sessions
7. Chat: knowledge-grounded Q&A
8. Message management: fetch and manage messages
9. Evaluation: evaluate model performance

## Detailed APIs

Below are the detailed API specs with examples.

### Tenant Management

| Method | Path            | Description            |
| ------ | --------------- | ---------------------- |
| POST   | `/tenants`      | Create a new tenant    |
| GET    | `/tenants/:id`  | Get a tenant by ID     |
| PUT    | `/tenants/:id`  | Update a tenant        |
| DELETE | `/tenants/:id`  | Delete a tenant        |
| GET    | `/tenants`      | List tenants           |

#### POST `/tenants` — Create a new tenant

Request:

```curl
curl --location 'http://localhost:8080/api/v1/tenants' \
--header 'Content-Type: application/json' \
--data '{
    "name": "weknorust",
    "description": "weknorust tenants",
    "business": "wechat",
    "retriever_engines": {
        "engines": [
            {
                "retriever_type": "keywords",
                "retriever_engine_type": "postgres"
            },
            {
                "retriever_type": "vector",
                "retriever_engine_type": "postgres"
            }
        ]
    }
}'
```

Response (example):

```json
{
  "data": {
    "id": 10000,
    "name": "weknorust",
    "description": "weknorust tenants",
    "api_key": "<redacted>",
    "status": "active",
    "retriever_engines": {
      "engines": [
        { "retriever_engine_type": "postgres", "retriever_type": "keywords" },
        { "retriever_engine_type": "postgres", "retriever_type": "vector" }
      ]
    },
    "business": "wechat",
    "storage_quota": 10737418240,
    "storage_used": 0,
    "created_at": "2025-08-11T20:37:28.396980093+08:00",
    "updated_at": "2025-08-11T20:37:28.396980301+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### GET `/tenants/:id` — Get tenant by ID

Request:

```curl
curl --location 'http://localhost:8080/api/v1/tenants/10000' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response (example):

```json
{
  "data": {
    "id": 10000,
    "name": "weknorust",
    "description": "weknorust tenants",
    "api_key": "<redacted>",
    "status": "active",
    "retriever_engines": {
      "engines": [
        { "retriever_engine_type": "postgres", "retriever_type": "keywords" },
        { "retriever_engine_type": "postgres", "retriever_type": "vector" }
      ]
    },
    "business": "wechat",
    "storage_quota": 10737418240,
    "storage_used": 0,
    "created_at": "2025-08-11T20:37:28.39698+08:00",
    "updated_at": "2025-08-11T20:37:28.405693+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### PUT `/tenants/:id` — Update tenant

Note: API key may be rotated.

Request:

```curl
curl --location --request PUT 'http://localhost:8080/api/v1/tenants/10000' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>' \
--data '{
  "name": "weknorust new",
  "description": "weknorust tenants new",
  "status": "active",
  "retriever_engines": {
    "engines": [
      { "retriever_engine_type": "postgres", "retriever_type": "keywords" },
      { "retriever_engine_type": "postgres", "retriever_type": "vector" }
    ]
  },
  "business": "wechat",
  "storage_quota": 10737418240
}'
```

Response (example):

```json
{
  "data": {
    "id": 10000,
    "name": "weknorust new",
    "description": "weknowust tenants new",
    "api_key": "<redacted>",
    "status": "active",
    "retriever_engines": {
      "engines": [
        { "retriever_engine_type": "postgres", "retriever_type": "keywords" },
        { "retriever_engine_type": "postgres", "retriever_type": "vector" }
      ]
    },
    "business": "wechat",
    "storage_quota": 10737418240,
    "storage_used": 0,
    "created_at": "0001-01-01T00:00:00Z",
    "updated_at": "2025-08-11T20:49:02.13421034+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### DELETE `/tenants/:id` — Delete tenant

Request:

```curl
curl --location --request DELETE 'http://localhost:8080/api/v1/tenants/10000' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response:

```json
{ "message": "Tenant deleted successfully", "success": true }
```

#### GET `/tenants` — List tenants

Request:

```curl
curl --location 'http://localhost:8080/api/v1/tenants' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response (example):

```json
{
  "data": {
    "items": [
      {
        "id": 10002,
        "name": "weknorust",
        "description": "weknorust tenants",
        "api_key": "<redacted>",
        "status": "active",
        "retriever_engines": {
          "engines": [
            { "retriever_engine_type": "postgres", "retriever_type": "keywords" },
            { "retriever_engine_type": "postgres", "retriever_type": "vector" }
          ]
        },
        "business": "wechat",
        "storage_quota": 10737418240,
        "storage_used": 0,
        "created_at": "2025-08-11T20:52:58.05679+08:00",
        "updated_at": "2025-08-11T20:52:58.060495+08:00",
        "deleted_at": null
      }
    ]
  },
  "success": true
}
```

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Knowledge Base Management

The following endpoints mirror those shown in the Chinese documentation (`docs/API.md`). Examples retain the same fields and shapes.

| Method | Path                                  | Description                         |
| ------ | ------------------------------------- | ----------------------------------- |
| POST   | `/knowledge-bases`                    | Create a knowledge base             |
| GET    | `/knowledge-bases`                    | List knowledge bases                |
| GET    | `/knowledge-bases/:id`                | Get knowledge base details          |
| PUT    | `/knowledge-bases/:id`                | Update a knowledge base             |
| DELETE | `/knowledge-bases/:id`                | Delete a knowledge base             |
| GET    | `/knowledge-bases/:id/hybrid-search`  | Hybrid search within a knowledge base |
| POST   | `/knowledge-bases/copy`               | Copy a knowledge base               |

Note: Chunking configuration fields include `chunk_size`, `chunk_overlap`, `separators`, and `enable_multimodal`.

#### POST `/knowledge-bases` — Create knowledge base

Request:

```curl
curl --location 'http://localhost:8080/api/v1/knowledge-bases' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>' \
--data '{
    "name": "weknorust",
    "description": "weknorust description",
    "chunking_config": {
        "chunk_size": 1000,
        "chunk_overlap": 200,
        "separators": [
            "."
        ],
        "enable_multimodal": true
    },
    "image_processing_config": {
        "model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e"
    },
    "embedding_model_id": "dff7bc94-7885-4dd1-bfd5-bd96e4df2fc3",
    "summary_model_id": "8aea788c-bb30-4898-809e-e40c14ffb48c",
    "rerank_model_id": "b30171a1-787b-426e-a293-735cd5ac16c0",
    "vlm_model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e",
    "vlm_config": {
        "model_name": "qwen2.5vl:3b",
        "interface_type": "ollama",
        "base_url": "",
        "api_key": ""
    },
    "cos_config": {
        "secret_id": "",
        "secret_key": "",
        "region": "",
        "bucket_name": "",
        "app_id": "",
        "path_prefix": ""
    }
}'
```

Response (example):

```json
{
  "data": {
    "id": "b5829e4a-3845-4624-a7fb-ea3b35e843b0",
    "name": "weknorust",
    "description": "weknorust description",
    "tenant_id": 1,
    "chunking_config": {
      "chunk_size": 1000,
      "chunk_overlap": 200,
      "separators": ["."],
      "enable_multimodal": true
    },
    "image_processing_config": { "model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e" },
    "embedding_model_id": "dff7bc94-7885-4dd1-bfd5-bd96e4df2fc3",
    "summary_model_id": "8aea788c-bb30-4898-809e-e40c14ffb48c",
    "rerank_model_id": "b30171a1-787b-426e-a293-735cd5ac16c0",
    "vlm_model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e",
    "vlm_config": {
      "model_name": "qwen2.5vl:3b",
      "base_url": "",
      "api_key": "",
      "interface_type": "ollama"
    },
    "cos_config": {
      "secret_id": "",
      "secret_key": "",
      "region": "",
      "bucket_name": "",
      "app_id": "",
      "path_prefix": ""
    },
    "created_at": "2025-08-12T11:30:09.206238645+08:00",
    "updated_at": "2025-08-12T11:30:09.206238854+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### GET `/knowledge-bases` — List knowledge bases

Request:

```curl
curl --location 'http://localhost:8080/api/v1/knowledge-bases' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response (example):

```json
{
  "data": [
    {
      "id": "kb-00000001",
      "name": "Default Knowledge Base",
      "description": "System Default Knowledge Base",
      "tenant_id": 1,
      "chunking_config": {
        "chunk_size": 1000,
        "chunk_overlap": 200,
        "separators": ["\n\n", "\n", "。", "！", "？", ";", "；"],
        "enable_multimodal": true
      },
      "image_processing_config": { "model_id": "" },
      "embedding_model_id": "dff7bc94-7885-4dd1-bfd5-bd96e4df2fc3",
      "summary_model_id": "8aea788c-bb30-4898-809e-e40c14ffb48c",
      "rerank_model_id": "b30171a1-787b-426e-a293-735cd5ac16c0",
      "vlm_model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e",
      "vlm_config": {
        "model_name": "qwen2.5vl:3b",
        "base_url": "http://host.docker.internal:11435/v1",
        "api_key": "",
        "interface_type": "ollama"
      },
      "cos_config": {
        "secret_id": "",
        "secret_key": "",
        "region": "",
        "bucket_name": "",
        "app_id": "",
        "path_prefix": ""
      },
      "created_at": "2025-08-11T20:10:41.817794+08:00",
      "updated_at": "2025-08-12T11:23:00.593097+08:00",
      "deleted_at": null
    }
  ],
  "success": true
}
```

#### GET `/knowledge-bases/:id` — Get knowledge base details

Request:

```curl
curl --location 'http://localhost:8080/api/v1/knowledge-bases/kb-00000001' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response (example):

```json
{
  "data": {
    "id": "kb-00000001",
    "name": "Default Knowledge Base",
    "description": "System Default Knowledge Base",
    "tenant_id": 1,
    "chunking_config": {
      "chunk_size": 1000,
      "chunk_overlap": 200,
      "separators": ["\n\n", "\n", "。", "！", "？", ";", "；"],
      "enable_multimodal": true
    },
    "image_processing_config": { "model_id": "" },
    "embedding_model_id": "dff7bc94-7885-4dd1-bfd5-bd96e4df2fc3",
    "summary_model_id": "8aea788c-bb30-4898-809e-e40c14ffb48c",
    "rerank_model_id": "b30171a1-787b-426e-a293-735cd5ac16c0",
    "vlm_model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e",
    "vlm_config": {
      "model_name": "qwen2.5vl:3b",
      "base_url": "http://host.docker.internal:11435/v1",
      "api_key": "",
      "interface_type": "ollama"
    },
    "cos_config": {
      "secret_id": "",
      "secret_key": "",
      "region": "",
      "bucket_name": "",
      "app_id": "",
      "path_prefix": ""
    },
    "created_at": "2025-08-11T20:10:41.817794+08:00",
    "updated_at": "2025-08-12T11:23:00.593097+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### PUT `/knowledge-bases/:id` — Update knowledge base

Request:

```curl
curl --location --request PUT 'http://localhost:8080/api/v1/knowledge-bases/b5829e4a-3845-4624-a7fb-ea3b35e843b0' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>' \
--data '{
  "name": "weknorust new",
  "description": "weknorust description new",
  "config": {
    "chunking_config": {
      "chunk_size": 1000,
      "chunk_overlap": 200,
      "separators": ["\n\n", "\n", "。", "！", "？", ";", "；"],
      "enable_multimodal": true
    },
    "image_processing_config": { "model_id": "" }
  }
}'
```

Response (example):

```json
{
  "data": {
    "id": "b5829e4a-3845-4624-a7fb-ea3b35e843b0",
    "name": "weknorust new",
    "description": "weknorust description new",
    "tenant_id": 1,
    "chunking_config": {
      "chunk_size": 1000,
      "chunk_overlap": 200,
      "separators": ["\n\n", "\n", "。", "！", "？", ";", "；"],
      "enable_multimodal": true
    },
    "image_processing_config": { "model_id": "" },
    "embedding_model_id": "dff7bc94-7885-4dd1-bfd5-bd96e4df2fc3",
    "summary_model_id": "8aea788c-bb30-4898-809e-e40c14ffb48c",
    "rerank_model_id": "b30171a1-787b-426e-a293-735cd5ac16c0",
    "vlm_model_id": "f2083ad7-63e3-486d-a610-e6c56e58d72e",
    "vlm_config": { "model_name": "qwen2.5vl:3b", "base_url": "", "api_key": "", "interface_type": "ollama" },
    "cos_config": { "secret_id": "", "secret_key": "", "region": "", "bucket_name": "", "app_id": "", "path_prefix": "" },
    "created_at": "2025-08-12T11:30:09.206238+08:00",
    "updated_at": "2025-08-12T11:36:09.083577609+08:00",
    "deleted_at": null
  },
  "success": true
}
```

#### DELETE `/knowledge-bases/:id` — Delete knowledge base

Request:

```curl
curl --location --request DELETE 'http://localhost:8080/api/v1/knowledge-bases/b5829e4a-3845-4624-a7fb-ea3b35e843b0' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>'
```

Response:

```json
{ "message": "Knowledge base deleted successfully", "success": true }
```

#### GET `/knowledge-bases/:id/hybrid-search` — Hybrid search in a knowledge base

Request:

```curl
curl --location --request GET 'http://localhost:8080/api/v1/knowledge-bases/kb-00000001/hybrid-search' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>' \
--data '{
  "query_text": "comet",
  "vector_threshold": 0.1,
  "keyword_threshold": 0.1,
  "match_count": 1
}'
```

Response (example):

```json
{
  "data": [
    {
      "id": "7d955251-3f79-4fd5-a6aa-02f81e044091",
      "content": "...",
      "knowledge_id": "a6790b93-4700-4676-bd48-0d4804e1456b",
      "chunk_index": 3,
      "knowledge_title": "彗星.txt",
      "start_at": 2287,
      "end_at": 2760,
      "seq": 3,
      "score": 0.7402352891601821,
      "match_type": 2,
      "sub_chunk_id": null,
      "metadata": {},
      "chunk_type": "text",
      "parent_chunk_id": "",
      "image_info": "",
      "knowledge_filename": "彗星.txt",
      "knowledge_source": ""
    }
  ],
  "success": true
}
```

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Knowledge Management

| Method | Path                                   | Description                         |
| ------ | -------------------------------------- | ----------------------------------- |
| POST   | `/knowledge-bases/:id/knowledge/file`  | Create knowledge from a file        |
| POST   | `/knowledge-bases/:id/knowledge/url`   | Create knowledge from a URL         |
| GET    | `/knowledge-bases/:id/knowledge`       | List knowledge under a base         |
| GET    | `/knowledge/:id`                       | Get knowledge details               |
| DELETE | `/knowledge/:id`                       | Delete a knowledge item             |
| GET    | `/knowledge/:id/download`              | Download a knowledge file           |
| PUT    | `/knowledge/:id`                       | Update knowledge metadata           |
| PUT    | `/knowledge/image/:id/:chunk_id`       | Update image chunk information      |
| GET    | `/knowledge/batch`                     | Batch get knowledge items           |

#### POST `/knowledge-bases/:id/knowledge/file` — Create knowledge from a file

Request:

```curl
curl --location 'http://localhost:8080/api/v1/knowledge-bases/kb-00000001/knowledge/file' \
--header 'Content-Type: application/json' \
--header 'X-API-Key: <your_api_key>' \
--form 'file=@"/path/to/彗星.txt"' \
--form 'enable_multimodel="true"'
```

Response: see `docs/API.md` example; fields are identical in English.

#### POST `/knowledge-bases/:id/knowledge/url` — Create knowledge from a URL

Request:

```curl
curl --location 'http://localhost:8080/api/v1/knowledge-bases/kb-00000001/knowledge/url' \
--header 'X-API-Key: <your_api_key>' \
--header 'Content-Type: application/json' \
--data '{
  "url":"https://github.com/SHA888/WeKnoRust",
  "enable_multimodel":true
}'
```

Response: identical to `docs/API.md` with English narrative.

#### GET `/knowledge-bases/:id/knowledge?page=&page_size` — List knowledge under a base

Request and response examples mirror `docs/API.md`.

#### GET `/knowledge/:id` — Get knowledge details

Request and response examples mirror `docs/API.md`.

#### GET `/knowledge/batch` — Batch get knowledge items

Request and response examples mirror `docs/API.md`.

#### DELETE `/knowledge/:id` — Delete knowledge

Response:

```json
{ "message": "Deleted successfully", "success": true }
```

#### GET `/knowledge/:id/download` — Download knowledge file

Response:

```
attachment
```

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Model Management

| Method | Path           | Description          |
| ------ | -------------- | -------------------- |
| POST   | `/models`      | Create a model       |
| GET    | `/models`      | List models          |
| GET    | `/models/:id`  | Get model details    |
| PUT    | `/models/:id`  | Update a model       |
| DELETE | `/models/:id`  | Delete a model       |

#### POST `/models` — Create models

- KnowledgeQA example request body mirrors `docs/API.md`.
- Embedding example request body mirrors `docs/API.md`.
- Rerank example request body mirrors `docs/API.md`.

Response examples for create/list/get/update/delete all mirror `docs/API.md`, with English narrative.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Chunk Management

| Method | Path                          | Description                       |
| ------ | ----------------------------- | --------------------------------- |
| GET    | `/chunks/:knowledge_id`       | List chunks for a knowledge item  |
| DELETE | `/chunks/:knowledge_id/:id`   | Delete a chunk                    |
| DELETE | `/chunks/:knowledge_id`       | Delete all chunks under knowledge |

Examples mirror `docs/API.md` with English descriptions.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Session Management

| Method | Path                                   | Description                     |
| ------ | -------------------------------------- | ------------------------------- |
| POST   | `/sessions`                             | Create a session                |
| GET    | `/sessions/:id`                         | Get session details             |
| GET    | `/sessions`                             | List sessions for a tenant      |
| PUT    | `/sessions/:id`                         | Update a session                |
| DELETE | `/sessions/:id`                         | Delete a session                |
| POST   | `/sessions/:session_id/generate_title`  | Generate session title          |
| GET    | `/sessions/continue-stream/:session_id` | Continue an unfinished session  |

Notes:
- `fallback_response` in examples contains Chinese text in `docs/API.md`; you may replace with English like "Sorry, I can’t answer that question." in your calls.
- Server-Sent Events streams are returned by streaming endpoints.

Examples mirror `docs/API.md` with English narrative.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Chat

| Method | Path                           | Description                 |
| ------ | ------------------------------ | --------------------------- |
| POST   | `/knowledge-chat/:session_id`  | Knowledge-grounded Q&A      |
| POST   | `/knowledge-search`            | Search within knowledge     |

`/knowledge-chat/:session_id` returns Server-Sent Events (Content-Type: text/event-stream). The SSE `message` events provide `references` and `answer` payloads as shown in `docs/API.md`.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Message Management

| Method | Path                          | Description                               |
| ------ | ----------------------------- | ----------------------------------------- |
| GET    | `/messages/:session_id/load`  | Load recent messages for a session        |
| DELETE | `/messages/:session_id/:id`   | Delete a message                          |

Query parameters for load:
- `before_time`: earliest `created_at` from the previous page; omit to fetch latest
- `limit`: page size (default 20)

Examples mirror `docs/API.md` with English narrative.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>

---

### Evaluation

| Method | Path           | Description           |
| ------ | -------------- | --------------------- |
| GET    | `/evaluation`  | Get evaluation task   |
| POST   | `/evaluation`  | Create evaluation task|

GET example mirrors `docs/API.md`. For POST, provide a task spec and capture the returned `task_id`, then use GET with `task_id` to poll progress.

<div align="right"><a href="#weknorust-api-documentation-en">Back to top ↑</a></div>
