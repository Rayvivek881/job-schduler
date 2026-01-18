# DAG Builders Guide

Comprehensive guide for using the DAG builder libraries to create job scheduler workflows.

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Python Usage](#python-usage)
5. [TypeScript Usage](#typescript-usage)
6. [VQL Filter Building](#vql-filter-building)
7. [DAG Patterns](#dag-patterns)
8. [Best Practices](#best-practices)
9. [Integration Examples](#integration-examples)

## Overview

The DAG Builders provide convenient utilities for creating DAG (Directed Acyclic Graph) definitions for the Job Scheduler API. They eliminate manual JSON construction and provide type-safe, validated DAG creation.

### Supported Workflows

- **Email Export Enrichment**: Enrich CSV files with email finder API
- **Contact Import**: Import contacts from CSV to database
- **Contact Export**: Export filtered contacts from database to CSV
- **Company Import**: Import companies from CSV to database
- **Company Export**: Export filtered companies from database to CSV

### Key Features

- **Type Safety**: Full TypeScript support with complete type definitions
- **Validation**: Pre-submission validation to catch errors early
- **VQL Builder**: Type-safe construction of complex filter queries
- **Flexible Configuration**: Custom variables, retry settings, and more
- **Unified Interface**: Consistent API across all workflow types

## Installation

### Python

No installation required - pure Python library with no external dependencies. Simply import:

```python
from dag_builders import create_contact_import_dag
```

### TypeScript

```bash
cd client-libs/typescript
npm install
npm run build
```

Or use in your TypeScript project:

```typescript
import { createContactImportDAG } from "./path/to/dag-builders";
```

## Quick Start

### Python Example

```python
from dag_builders.contact_import import create_contact_import_dag

# Create DAG
dag = create_contact_import_dag(
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv"
)

# Submit to API
import requests
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

### TypeScript Example

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

// Create DAG
const dag = createContactImportDAG("my-bucket", "uploads/contacts.csv");

// Submit to API
const response = await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key",
  },
  body: JSON.stringify(dag),
});
```

## Python Usage

### Contact Import

```python
from dag_builders.contact_import import create_contact_import_dag

dag = create_contact_import_dag(
    s3_bucket="my-uploads-bucket",
    s3_key="uploads/contacts_2024.csv",
    retry_count=5,
    custom_vars={"source": "manual_upload"}
)
```

### Contact Export with Filters

```python
from dag_builders.contact_export import create_contact_export_dag
from utils.vql_builder import VQLFilterBuilder

# Build keyword filter
keyword_filter = VQLFilterBuilder.keyword_match(
    must={"country": ["united states"], "title": ["VP", "Director"]}
)

where = VQLFilterBuilder.combine_filters(keyword_match=keyword_filter)

dag = create_contact_export_dag(
    s3_bucket="my-exports-bucket",
    select_columns=["first_name", "last_name", "email", "title"],
    where=where,
    limit=1000
)
```

### Email Export Enrichment

```python
from dag_builders.email_export import create_email_export_dag

dag = create_email_export_dag(
    s3_bucket="my-exports-bucket",
    s3_key="exports/contacts.csv",
    email_api_url="https://api.emailfinder.com/v1",
    email_api_key="api-key-123",
    batch_size=100
)
```

### Unified Builder

```python
from dag_builders.unified import ImportExportDAGBuilder

# Import
dag = ImportExportDAGBuilder.create_import_dag(
    service="contact",
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv"
)

# Export
dag = ImportExportDAGBuilder.create_export_dag(
    service="contact",
    s3_bucket="my-bucket",
    select_columns=["first_name", "last_name", "email"],
    where=where_clause
)
```

## TypeScript Usage

### Contact Import

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

const dag = createContactImportDAG("my-uploads-bucket", "uploads/contacts_2024.csv", {
  retryCount: 5,
  customVars: { source: "manual_upload" },
});
```

### Contact Export with Filters

```typescript
import { createContactExportDAG } from "./src/builders/ContactExportDAGBuilder";
import { VQLFilterBuilder } from "./src/utils/VQLBuilder";

const keywordFilter = VQLFilterBuilder.keywordMatch({
  must: { country: ["united states"], title: ["VP", "Director"] },
});

const where = VQLFilterBuilder.combineFilters({ keywordMatch: keywordFilter });

const dag = createContactExportDAG(
  "my-exports-bucket",
  ["first_name", "last_name", "email", "title"],
  { where, limit: 1000 }
);
```

### Email Export Enrichment

```typescript
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";

const dag = createEmailExportDAG("my-exports-bucket", "exports/contacts.csv", {
  emailApiUrl: "https://api.emailfinder.com/v1",
  emailApiKey: "api-key-123",
  batchSize: 100,
});
```

## VQL Filter Building

### Keyword Matching

**Python:**
```python
from utils.vql_builder import VQLFilterBuilder

# Single field
keyword_filter = VQLFilterBuilder.keyword_match(
    must={"country": ["united states"]}
)

# Multiple fields
keyword_filter = VQLFilterBuilder.keyword_match(
    must={
        "country": ["united states"],
        "title": ["VP", "Director", "Manager"]
    },
    must_not={"status": ["inactive"]}
)
```

**TypeScript:**
```typescript
import { VQLFilterBuilder } from "./src/utils/VQLBuilder";

const keywordFilter = VQLFilterBuilder.keywordMatch({
  must: { country: ["united states"], title: ["VP", "Director"] },
  mustNot: { status: ["inactive"] },
});
```

### Text Matching

**Python:**
```python
text_matches = [
    VQLFilterBuilder.text_match(
        text_value="software engineer",
        filter_key="title",
        search_type="phrase",
        fuzzy=True
    )
]

where = VQLFilterBuilder.combine_filters(text_matches=text_matches)
```

**TypeScript:**
```typescript
const textMatches = [
  VQLFilterBuilder.textMatch({
    textValue: "software engineer",
    filterKey: "title",
    searchType: "phrase",
    fuzzy: true,
  }),
];

const where = VQLFilterBuilder.combineFilters({ textMatches });
```

### Combining Multiple Filters

**Python:**
```python
keyword_filter = VQLFilterBuilder.keyword_match(must={"country": ["united states"]})
text_matches = [VQLFilterBuilder.text_match(text_value="engineer", filter_key="title")]

where = VQLFilterBuilder.combine_filters(
    keyword_match=keyword_filter,
    text_matches=text_matches
)
```

**TypeScript:**
```typescript
const keywordFilter = VQLFilterBuilder.keywordMatch({
  must: { country: ["united states"] },
});

const textMatches = [
  VQLFilterBuilder.textMatch({
    textValue: "engineer",
    filterKey: "title",
  }),
];

const where = VQLFilterBuilder.combineFilters({
  keywordMatch: keywordFilter,
  textMatches,
});
```

## DAG Patterns

### Batch Import

**Python:**
```python
import_files = [
    "uploads/batch_1.csv",
    "uploads/batch_2.csv",
    "uploads/batch_3.csv",
]

dags = []
for s3_key in import_files:
    dag = create_contact_import_dag(
        s3_bucket="my-bucket",
        s3_key=s3_key,
        custom_vars={"batch_id": f"batch-{s3_key}"}
    )
    dags.extend(dag)

# Submit all at once (if < 5000 nodes)
response = requests.post(api_url, json=dags, headers=headers)
```

**TypeScript:**
```typescript
const importFiles = ["uploads/batch_1.csv", "uploads/batch_2.csv", "uploads/batch_3.csv"];

const dags: any[] = [];
for (const s3Key of importFiles) {
  const dag = createContactImportDAG("my-bucket", s3Key, {
    customVars: { batch_id: `batch-${s3Key}` },
  });
  dags.push(...dag);
}

// Submit all at once (if < 5000 nodes)
await fetch(apiUrl, {
  method: "POST",
  body: JSON.stringify(dags),
  headers: { "X-API-Key": apiKey },
});
```

### Complex Export with Multiple Filters

**Python:**
```python
from dag_builders.contact_export import create_contact_export_dag
from utils.vql_builder import VQLFilterBuilder

# Build complex filter
keyword_filter = VQLFilterBuilder.keyword_match(
    must={"country": ["united states"], "seniority": ["senior"]},
    must_not={"title": ["intern"]}
)

text_matches = [
    VQLFilterBuilder.text_match(
        text_value="engineering",
        filter_key="department",
        search_type="match"
    )
]

where = VQLFilterBuilder.combine_filters(
    keyword_match=keyword_filter,
    text_matches=text_matches
)

dag = create_contact_export_dag(
    s3_bucket="my-exports-bucket",
    select_columns=["first_name", "last_name", "email", "title", "company_name"],
    where=where,
    limit=2000,
    order_by=[
        {"order_by": "created_at", "order_direction": "desc"},
        {"order_by": "email", "order_direction": "asc"}
    ]
)
```

## Best Practices

### 1. Always Validate Before Submission

**Python:**
```python
from utils.validators import DAGValidator

dag = create_contact_import_dag(...)

is_valid, error = DAGValidator.validate_dag(dag)
if not is_valid:
    raise ValueError(f"DAG validation failed: {error}")
```

**TypeScript:**
```typescript
import { DAGValidator } from "./src/utils/Validators";

const dag = createContactImportDAG(...);

const validation = DAGValidator.validateDAG(dag);
if (!validation.isValid) {
  throw new Error(`DAG validation failed: ${validation.error}`);
}
```

### 2. Use Custom Variables for Metadata

```python
dag = create_contact_import_dag(
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv",
    custom_vars={
        "user_id": "user-123",
        "campaign_id": "campaign-456",
        "source": "api_upload"
    }
)
```

### 3. Configure Retry Settings Appropriately

- **Import jobs**: 3-5 retries, 5-10 minute intervals
- **Export jobs**: 2-3 retries, 3-5 minute intervals
- **Email enrichment**: 5+ retries, 10+ minute intervals (external API dependency)

### 4. Always Include UUID in Export Order By

The builders automatically add UUID to `order_by` for stable pagination. If you provide custom `order_by`, ensure UUID is included or it will be prepended automatically.

### 5. Use Select Columns Wisely

Export jobs require `select_columns`. Only select what you need to reduce export time and file size.

## Integration Examples

### Python Flask API

```python
from flask import Flask, request, jsonify
from dag_builders.contact_import import create_contact_import_dag

app = Flask(__name__)

@app.route("/api/create-import-job", methods=["POST"])
def create_import_job():
    data = request.json
    dag = create_contact_import_dag(
        s3_bucket=data["bucket"],
        s3_key=data["key"],
        custom_vars={"api_source": "flask"}
    )
    
    # Submit to job scheduler
    response = requests.post(
        "http://job-scheduler:8000/jobs/bulk-insert/complete-graph",
        json=dag,
        headers={"X-API-Key": os.getenv("JOB_SCHEDULER_API_KEY")}
    )
    
    return jsonify(response.json()), response.status_code
```

### TypeScript Express API

```typescript
import express from "express";
import { createContactImportDAG } from "./dag-builders";

const app = express();

app.post("/api/create-import-job", async (req, res) => {
  const { bucket, key } = req.body;
  
  const dag = createContactImportDAG(bucket, key, {
    customVars: { api_source: "express" },
  });
  
  // Submit to job scheduler
  const response = await fetch(
    "http://job-scheduler:8000/jobs/bulk-insert/complete-graph",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-API-Key": process.env.JOB_SCHEDULER_API_KEY!,
      },
      body: JSON.stringify(dag),
    }
  );
  
  const result = await response.json();
  res.status(response.status).json(result);
});
```

## Additional Resources

- [API Documentation](../docs/API.md) - Complete API reference
- [DAG Guide](../docs/DAG_GUIDE.md) - Understanding DAGs and dependencies
- [Python Examples](../client-libs/python/examples/) - Working examples
- [TypeScript Examples](../client-libs/typescript/examples/) - Working examples

## Troubleshooting

### Validation Errors

If DAG validation fails, check:
- All required fields are present (uuid, job_title, job_type, data)
- No duplicate UUIDs
- All edge targets exist in DAG
- Node count doesn't exceed 5000

### Type Errors (TypeScript)

Ensure you're using the correct types from `dag.types.ts` and `vql.types.ts`. The builders return `DAGDefinition` which is `JobNode[]`.

### Import Errors (Python)

Make sure the `dag_builders` package is in your Python path. You may need to add the parent directory:

```python
import sys
from pathlib import Path
sys.path.insert(0, str(Path(__file__).parent.parent))
```
