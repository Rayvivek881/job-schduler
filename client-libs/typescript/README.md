# TypeScript DAG Builders for Job Scheduler

TypeScript utilities for building DAG (Directed Acyclic Graph) definitions for the job scheduler system.

## Overview

This library provides convenient functions and classes for creating DAG payloads for the Job Scheduler API. It supports:

- **Email Export DAGs**: Enrich CSV files with email finder API
- **Contact Import/Export DAGs**: Import contacts from CSV or export filtered contacts
- **Company Import/Export DAGs**: Import companies from CSV or export filtered companies
- **VQL Filter Builders**: Type-safe construction of VQL (Vivek Query Language) filters
- **Validation Utilities**: Pre-submission DAG validation
- **Full TypeScript Support**: Complete type definitions for all structures

## Installation

```bash
npm install
npm run build
```

Or use directly in TypeScript projects:

```typescript
import { createContactImportDAG } from "@contact360/job-scheduler-dag-builders";
```

## Quick Start

### Contact Import

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

const dag = createContactImportDAG(
  "my-bucket",
  "uploads/contacts.csv"
);

// Submit to POST /jobs/bulk-insert/complete-graph
```

### Contact Export with Filters

```typescript
import { createContactExportDAG } from "./src/builders/ContactExportDAGBuilder";
import { VQLFilterBuilder } from "./src/utils/VQLBuilder";

// Build keyword filter
const keywordFilter = VQLFilterBuilder.keywordMatch({
  must: {
    country: ["united states"],
    title: ["VP", "Director"],
  },
});
const where = VQLFilterBuilder.combineFilters({ keywordMatch: keywordFilter });

// Create export DAG
const dag = createContactExportDAG(
  "my-exports-bucket",
  ["first_name", "last_name", "email", "title"],
  {
    where,
    limit: 1000,
  }
);
```

### Email Export Enrichment

```typescript
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";

const dag = createEmailExportDAG(
  "my-bucket",
  "exports/contacts.csv",
  {
    emailApiUrl: "https://api.emailfinder.com/v1",
    emailApiKey: "api-key-123",
    batchSize: 100,
  }
);
```

### Unified Builder

```typescript
import { ImportExportDAGBuilder } from "./src/builders/UnifiedDAGBuilder";

// Import
const dag = ImportExportDAGBuilder.createImportDAG(
  "contact",
  "my-bucket",
  "uploads/contacts.csv"
);

// Export
const dag = ImportExportDAGBuilder.createExportDAG(
  "contact",
  "my-bucket",
  ["first_name", "last_name", "email"],
  { where: whereClause }
);
```

## API Reference

### DAG Builders

#### `createContactImportDAG(s3Bucket, s3Key, options?)`
Create DAG for importing contacts from CSV file.

#### `createContactExportDAG(s3Bucket, selectColumns, options?)`
Create DAG for exporting filtered contacts to CSV.

#### `createCompanyImportDAG(s3Bucket, s3Key, options?)`
Create DAG for importing companies from CSV file.

#### `createCompanyExportDAG(s3Bucket, selectColumns, options?)`
Create DAG for exporting filtered companies to CSV.

#### `createEmailExportDAG(s3Bucket, s3Key, options?)`
Create DAG for email enrichment workflow.

### VQL Filter Builder

#### `VQLFilterBuilder.keywordMatch(options)`
Build keyword matching filter.

#### `VQLFilterBuilder.textMatch(options)`
Build text matching filter with fuzzy support.

#### `VQLFilterBuilder.rangeQuery(options)`
Build range query filter for numeric/date ranges.

#### `VQLFilterBuilder.buildVQLQuery(options)`
Build complete VQL query structure.

#### `VQLFilterBuilder.combineFilters(options)`
Combine multiple filter types into where clause.

### Validation

#### `DAGValidator.validateDAG(dag)`
Validate DAG structure before submission.

## Examples

See the `examples/` directory for complete usage examples:

- `emailExport.example.ts` - Email enrichment workflows
- `contactImport.example.ts` - Contact import examples
- `batchOperations.example.ts` - Batch operations and complex filters

## Type Safety

All functions and structures are fully typed. TypeScript will provide autocomplete and type checking for:

- DAG definitions
- VQL queries
- Job data structures
- Filter options

## Integration with Job Scheduler API

Once you have a DAG, submit it to the job scheduler:

```typescript
const dag = createContactImportDAG(...);

const response = await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key",
  },
  body: JSON.stringify(dag),
});

const result = await response.json();
```

## Building

```bash
# Build TypeScript
npm run build

# Watch mode
npm run watch

# Clean build artifacts
npm run clean
```

## License

Part of the Job Scheduler project.
