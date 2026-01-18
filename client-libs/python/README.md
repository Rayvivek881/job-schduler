# Python DAG Builders for Job Scheduler

Python utilities for building DAG (Directed Acyclic Graph) definitions for the job scheduler system.

## Overview

This library provides convenient functions and classes for creating DAG payloads for the Job Scheduler API. It supports:

- **Email Export DAGs**: Enrich CSV files with email finder API
- **Contact Import/Export DAGs**: Import contacts from CSV or export filtered contacts
- **Company Import/Export DAGs**: Import companies from CSV or export filtered companies
- **VQL Filter Builders**: Type-safe construction of VQL (Vivek Query Language) filters
- **Validation Utilities**: Pre-submission DAG validation
- **Template System**: Reusable DAG patterns

## Installation

No installation required - this is a pure Python library with no external dependencies. All functionality uses Python standard library only.

Simply import the modules:

```python
from dag_builders import create_contact_import_dag
from utils.vql_builder import VQLFilterBuilder
```

## Quick Start

### Contact Import

```python
from dag_builders.contact_import import create_contact_import_dag

dag = create_contact_import_dag(
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv",
)

# Submit to POST /jobs/bulk-insert/complete-graph
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

# Create export DAG
dag = create_contact_export_dag(
    s3_bucket="my-exports-bucket",
    select_columns=["first_name", "last_name", "email", "title"],
    where=where,
    limit=1000,
)
```

### Email Export Enrichment

```python
from dag_builders.email_export import create_email_export_dag

dag = create_email_export_dag(
    s3_bucket="my-bucket",
    s3_key="exports/contacts.csv",
    email_api_url="https://api.emailfinder.com/v1",
    email_api_key="api-key-123",
    batch_size=100,
)
```

### Unified Builder

```python
from dag_builders.unified import ImportExportDAGBuilder

# Import
dag = ImportExportDAGBuilder.create_import_dag(
    service="contact",
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv",
)

# Export
dag = ImportExportDAGBuilder.create_export_dag(
    service="contact",
    s3_bucket="my-bucket",
    select_columns=["first_name", "last_name", "email"],
    where=where_clause,
)
```

## API Reference

### DAG Builders

#### `create_contact_import_dag()`
Create DAG for importing contacts from CSV file.

#### `create_contact_export_dag()`
Create DAG for exporting filtered contacts to CSV.

#### `create_company_import_dag()`
Create DAG for importing companies from CSV file.

#### `create_company_export_dag()`
Create DAG for exporting filtered companies to CSV.

#### `create_email_export_dag()`
Create DAG for email enrichment workflow.

### VQL Filter Builder

#### `VQLFilterBuilder.keyword_match()`
Build keyword matching filter.

#### `VQLFilterBuilder.text_match()`
Build text matching filter with fuzzy support.

#### `VQLFilterBuilder.range_query()`
Build range query filter for numeric/date ranges.

#### `VQLFilterBuilder.build_vql_query()`
Build complete VQL query structure.

#### `VQLFilterBuilder.combine_filters()`
Combine multiple filter types into where clause.

### Validation

#### `DAGValidator.validate_dag()`
Validate DAG structure before submission.

## Examples

See the `examples/` directory for complete usage examples:

- `email_export_example.py` - Email enrichment workflows
- `contact_import_example.py` - Contact import examples
- `batch_operations_example.py` - Batch operations and complex filters

## Data Structures

### Job Data Structure

Jobs created by these builders follow the standard job scheduler format:

```json
{
  "uuid": "contact-import-20240115-120000-abc123",
  "job_title": "Import Contacts from CSV: uploads/contacts.csv",
  "job_type": "insert_csv_file",
  "data": {
    "s3_key": "uploads/contacts.csv",
    "s3_bucket": "my-bucket",
    "service": "contact",
    "workflow_uuid": "...",
    "custom_vars": { ... }
  },
  "retry_count": 3,
  "retry_interval": 5,
  "edges": []
}
```

### VQL Query Structure

VQL queries for exports follow this structure:

```json
{
  "where": {
    "keyword_match": {
      "must": { "country": ["united states"] },
      "must_not": {}
    },
    "text_matches": {
      "must": [],
      "must_not": []
    },
    "range_query": {}
  },
  "select_columns": ["first_name", "last_name", "email"],
  "order_by": [{"order_by": "uuid", "order_direction": "desc"}],
  "limit": 1000
}
```

## Integration with Job Scheduler API

Once you have a DAG, submit it to the job scheduler:

```python
import requests

dag = create_contact_import_dag(...)

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={
        "Content-Type": "application/json",
        "X-API-Key": "your-api-key",
    },
    json=dag,
)
response.raise_for_status()
```

## License

Part of the Job Scheduler project.
