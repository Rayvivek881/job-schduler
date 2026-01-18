# Complete Guide: Creating Dynamic DAGs with Custom Variables

This comprehensive guide explains how to create dynamic DAGs (Directed Acyclic Graphs) with custom variables for flexible, data-driven job scheduling.

---

## 📚 Table of Contents

1. [Understanding Dynamic DAGs](#understanding-dynamic-dags)
2. [Custom Variables Overview](#custom-variables-overview)
3. [Dynamic DAG Creation Patterns](#dynamic-dag-creation-patterns)
4. [Using Custom Variables in Builders](#using-custom-variables-in-builders)
5. [Dynamic DAG Patterns](#dynamic-dag-patterns)
6. [Template-Based DAG Creation](#template-based-dag-creation)
7. [Querying Jobs by Custom Variables](#querying-jobs-by-custom-variables)
8. [Complete Examples](#complete-examples)
9. [Best Practices](#best-practices)
10. [Task Breakdown](#task-breakdown)

---

## Understanding Dynamic DAGs

### What are Dynamic DAGs?

**Dynamic DAGs** are DAGs created programmatically based on runtime data, as opposed to static DAGs defined at compile time.

### Key Characteristics

| Feature | Static DAG | Dynamic DAG |
|---------|-----------|-------------|
| **Creation** | Hard-coded in source code | Generated at runtime |
| **Variables** | Fixed values | Runtime values |
| **Flexibility** | Limited | High |
| **Use Cases** | Simple workflows | Batch processing, scheduled jobs |

### Examples of Dynamic Scenarios

1. **Batch Processing**: Create DAGs for multiple files based on S3 file list
2. **Scheduled Jobs**: Generate DAGs daily/weekly with date-based variables
3. **Multi-tenant**: Create DAGs per user/organization with tenant IDs
4. **Configurable Workflows**: Build DAGs from configuration files
5. **Runtime Data**: Generate DAGs from database queries or API responses

---

## Custom Variables Overview

### What are Custom Variables?

**Custom Variables** are metadata stored in the `data.custom_vars` field of each job, allowing you to:
- Add context to jobs (user_id, campaign_id, date, etc.)
- Query and filter jobs by metadata
- Track job origin and purpose
- Group related jobs together

### Custom Variables Structure

```json
{
  "data": {
    "custom_vars": {
      "user_id": "user-123",
      "campaign_id": "campaign-456",
      "date": "2024-01-15",
      "region": "us-east-1",
      "source": "api_upload",
      "batch_id": "batch-2024-01-15-001"
    }
  }
}
```

### Benefits

1. **Querying**: Filter jobs by custom variables (stored as JSONB in PostgreSQL)
2. **Grouping**: Find all jobs from same campaign/user/date
3. **Tracking**: Monitor jobs by origin or purpose
4. **Filtering**: Create dynamic filters based on metadata

---

## Dynamic DAG Creation Patterns

### Pattern 1: Loop-Based Generation

Create multiple jobs from a list of items.

**Python Example:**
```python
from dag_builders.contact_import import create_contact_import_dag

# List of files to process
import_files = [
    "uploads/contacts_batch_1.csv",
    "uploads/contacts_batch_2.csv",
    "uploads/contacts_batch_3.csv",
]

dags = []
for s3_key in import_files:
    dag = create_contact_import_dag(
        s3_bucket="my-uploads-bucket",
        s3_key=s3_key,
        custom_vars={
            "batch_id": f"batch-{s3_key.split('_')[-1].split('.')[0]}",
            "source": "batch_upload",
            "upload_date": "2024-01-15"
        }
    )
    dags.extend(dag)

# Submit all DAGs at once
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dags
)
```

**TypeScript Example:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

const importFiles = [
  "uploads/contacts_batch_1.csv",
  "uploads/contacts_batch_2.csv",
  "uploads/contacts_batch_3.csv",
];

const dags: any[] = [];
for (const s3Key of importFiles) {
  const dag = createContactImportDAG("my-uploads-bucket", s3Key, {
    customVars: {
      batch_id: `batch-${s3Key.split("_").pop()?.split(".")[0]}`,
      source: "batch_upload",
      upload_date: "2024-01-15"
    }
  });
  dags.push(...dag);
}

// Submit all DAGs at once
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dags)
});
```

### Pattern 2: Date-Based Dynamic DAGs

Create DAGs with date-specific variables.

**Python Example:**
```python
from datetime import datetime, timedelta
from dag_builders.contact_import import create_contact_import_dag

def create_daily_import_dag(date_str: str, bucket: str):
    """Create import DAG for a specific date"""
    s3_key = f"daily_uploads/{date_str}/contacts.csv"
    
    dag = create_contact_import_dag(
        s3_bucket=bucket,
        s3_key=s3_key,
        custom_vars={
            "date": date_str,
            "source": "daily_upload",
            "upload_type": "scheduled",
            "region": "us-east-1"
        }
    )
    return dag

# Create DAGs for last 7 days
dags = []
for i in range(7):
    date = (datetime.now() - timedelta(days=i)).strftime("%Y-%m-%d")
    dag = create_daily_import_dag(date, "my-bucket")
    dags.extend(dag)

# Submit all DAGs
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dags
)
```

**TypeScript Example:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

function createDailyImportDAG(dateStr: string, bucket: string) {
  const s3Key = `daily_uploads/${dateStr}/contacts.csv`;
  
  return createContactImportDAG(bucket, s3Key, {
    customVars: {
      date: dateStr,
      source: "daily_upload",
      upload_type: "scheduled",
      region: "us-east-1"
    }
  });
}

// Create DAGs for last 7 days
const dags: any[] = [];
for (let i = 0; i < 7; i++) {
  const date = new Date();
  date.setDate(date.getDate() - i);
  const dateStr = date.toISOString().split("T")[0];
  const dag = createDailyImportDAG(dateStr, "my-bucket");
  dags.push(...dag);
}

// Submit all DAGs
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dags)
});
```

### Pattern 3: Config-Driven Dynamic DAGs

Create DAGs from configuration dictionaries.

**Python Example:**
```python
from dag_builders.contact_import import create_contact_import_dag
from dag_builders.base import BaseDAGBuilder

def create_etl_pipeline_dag(date_str: str, bucket: str, config: dict):
    """Create ETL pipeline DAG dynamically with custom variables"""
    
    base_uuid_prefix = f"etl-{date_str}"
    
    dag = [
        {
            "uuid": f"{base_uuid_prefix}-extract",
            "job_title": f"Extract Data - {date_str}",
            "job_type": "etl",
            "data": {
                "source_bucket": bucket,
                "source_key": f"data/{date_str}/input.csv",
                "format": config.get("format", "csv"),
                "custom_vars": {
                    "date": date_str,
                    "region": config.get("region", "us-east-1"),
                    "workflow_type": "etl_pipeline",
                    "pipeline_id": config.get("pipeline_id", "default")
                }
            },
            "retry_count": config.get("retry_count", 3),
            "retry_interval": config.get("retry_interval", 5),
            "edges": [f"{base_uuid_prefix}-transform"]
        },
        {
            "uuid": f"{base_uuid_prefix}-transform",
            "job_title": f"Transform Data - {date_str}",
            "job_type": "etl",
            "data": {
                "transformations": config.get("transformations", []),
                "output_format": config.get("output_format", "parquet"),
                "custom_vars": {
                    "date": date_str,
                    "region": config.get("region", "us-east-1"),
                    "workflow_type": "etl_pipeline",
                    "pipeline_id": config.get("pipeline_id", "default")
                }
            },
            "retry_count": config.get("retry_count", 3),
            "retry_interval": config.get("retry_interval", 5),
            "edges": [f"{base_uuid_prefix}-load"]
        },
        {
            "uuid": f"{base_uuid_prefix}-load",
            "job_title": f"Load to Warehouse - {date_str}",
            "job_type": "etl",
            "data": {
                "destination": config.get("warehouse", "warehouse"),
                "table": config.get("table", "daily_data"),
                "custom_vars": {
                    "date": date_str,
                    "region": config.get("region", "us-east-1"),
                    "workflow_type": "etl_pipeline",
                    "pipeline_id": config.get("pipeline_id", "default")
                }
            },
            "edges": []
        }
    ]
    
    return dag

# Usage
date_str = "2024-01-15"
config = {
    "format": "csv",
    "transformations": ["clean", "normalize"],
    "output_format": "parquet",
    "warehouse": "data_warehouse",
    "table": "daily_processed",
    "region": "us-east-1",
    "pipeline_id": "pipeline-001",
    "retry_count": 5,
    "retry_interval": 10
}

dag = create_etl_pipeline_dag(date_str, "my-bucket", config)

response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dag
)
```

**TypeScript Example:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

interface ETLConfig {
  format?: string;
  transformations?: string[];
  output_format?: string;
  warehouse?: string;
  table?: string;
  region?: string;
  pipeline_id?: string;
  retry_count?: number;
  retry_interval?: number;
}

function createETLPipelineDAG(dateStr: string, bucket: string, config: ETLConfig) {
  const baseUuidPrefix = `etl-${dateStr}`;
  
  const dag = [
    BaseDAGBuilder.buildJobNode(
      `${baseUuidPrefix}-extract`,
      `Extract Data - ${dateStr}`,
      "etl",
      {
        source_bucket: bucket,
        source_key: `data/${dateStr}/input.csv`,
        format: config.format || "csv",
        custom_vars: {
          date: dateStr,
          region: config.region || "us-east-1",
          workflow_type: "etl_pipeline",
          pipeline_id: config.pipeline_id || "default"
        }
      },
      {
        edges: [`${baseUuidPrefix}-transform`],
        retryCount: config.retry_count || 3,
        retryInterval: config.retry_interval || 5
      }
    ),
    BaseDAGBuilder.buildJobNode(
      `${baseUuidPrefix}-transform`,
      `Transform Data - ${dateStr}`,
      "etl",
      {
        transformations: config.transformations || [],
        output_format: config.output_format || "parquet",
        custom_vars: {
          date: dateStr,
          region: config.region || "us-east-1",
          workflow_type: "etl_pipeline",
          pipeline_id: config.pipeline_id || "default"
        }
      },
      {
        edges: [`${baseUuidPrefix}-load`],
        retryCount: config.retry_count || 3,
        retryInterval: config.retry_interval || 5
      }
    ),
    BaseDAGBuilder.buildJobNode(
      `${baseUuidPrefix}-load`,
      `Load to Warehouse - ${dateStr}`,
      "etl",
      {
        destination: config.warehouse || "warehouse",
        table: config.table || "daily_data",
        custom_vars: {
          date: dateStr,
          region: config.region || "us-east-1",
          workflow_type: "etl_pipeline",
          pipeline_id: config.pipeline_id || "default"
        }
      },
      {
        edges: [],
        retryCount: config.retry_count || 3,
        retryInterval: config.retry_interval || 5
      }
    )
  ];
  
  return dag;
}

// Usage
const dateStr = "2024-01-15";
const config: ETLConfig = {
  format: "csv",
  transformations: ["clean", "normalize"],
  output_format: "parquet",
  warehouse: "data_warehouse",
  table: "daily_processed",
  region: "us-east-1",
  pipeline_id: "pipeline-001",
  retry_count: 5,
  retry_interval: 10
};

const dag = createETLPipelineDAG(dateStr, "my-bucket", config);

await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dag)
});
```

### Pattern 4: Multi-Tenant Dynamic DAGs

Create DAGs per user/organization with tenant isolation.

**Python Example:**
```python
from dag_builders.contact_import import create_contact_import_dag

def create_user_import_dag(user_id: str, org_id: str, s3_key: str):
    """Create import DAG for specific user/organization"""
    
    dag = create_contact_import_dag(
        s3_bucket="user-uploads-bucket",
        s3_key=s3_key,
        custom_vars={
            "user_id": user_id,
            "org_id": org_id,
            "tenant_id": org_id,  # For multi-tenant isolation
            "source": "user_upload",
            "upload_type": "manual",
            "created_by": user_id
        }
    )
    return dag

# Process uploads for multiple users
user_uploads = [
    {"user_id": "user-123", "org_id": "org-001", "s3_key": "uploads/user-123/contacts.csv"},
    {"user_id": "user-456", "org_id": "org-002", "s3_key": "uploads/user-456/contacts.csv"},
    {"user_id": "user-789", "org_id": "org-001", "s3_key": "uploads/user-789/contacts.csv"},
]

dags = []
for upload in user_uploads:
    dag = create_user_import_dag(
        upload["user_id"],
        upload["org_id"],
        upload["s3_key"]
    )
    dags.extend(dag)

# Submit all DAGs
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dags
)
```

**TypeScript Example:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

interface UserUpload {
  userId: string;
  orgId: string;
  s3Key: string;
}

function createUserImportDAG(userId: string, orgId: string, s3Key: string) {
  return createContactImportDAG("user-uploads-bucket", s3Key, {
    customVars: {
      user_id: userId,
      org_id: orgId,
      tenant_id: orgId,  // For multi-tenant isolation
      source: "user_upload",
      upload_type: "manual",
      created_by: userId
    }
  });
}

// Process uploads for multiple users
const userUploads: UserUpload[] = [
  { userId: "user-123", orgId: "org-001", s3Key: "uploads/user-123/contacts.csv" },
  { userId: "user-456", orgId: "org-002", s3Key: "uploads/user-456/contacts.csv" },
  { userId: "user-789", orgId: "org-001", s3Key: "uploads/user-789/contacts.csv" },
];

const dags: any[] = [];
for (const upload of userUploads) {
  const dag = createUserImportDAG(upload.userId, upload.orgId, upload.s3Key);
  dags.push(...dag);
}

// Submit all DAGs
await fetch("http://localhost:8000/jobs/bulk-insert/complete-graph", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-API-Key": "your-api-key"
  },
  body: JSON.stringify(dags)
});
```

---

## Using Custom Variables in Builders

### Python: Adding Custom Variables

All Python builders accept `custom_vars` parameter:

```python
from dag_builders.contact_import import create_contact_import_dag
from dag_builders.email_export import create_email_export_dag
from dag_builders.contact_export import create_contact_export_dag

# Contact Import
dag = create_contact_import_dag(
    s3_bucket="my-bucket",
    s3_key="uploads/contacts.csv",
    custom_vars={
        "user_id": "user-123",
        "campaign_id": "campaign-456",
        "source": "manual_upload",
        "priority": "high"
    }
)

# Email Export
dag = create_email_export_dag(
    s3_bucket="my-bucket",
    s3_key="exports/contacts.csv",
    custom_vars={
        "user_id": "user-123",
        "campaign_id": "campaign-456",
        "export_type": "email_enrichment"
    }
)

# Contact Export
dag = create_contact_export_dag(
    s3_bucket="my-bucket",
    select_columns=["first_name", "last_name", "email"],
    custom_vars={
        "user_id": "user-123",
        "campaign_id": "campaign-456",
        "export_type": "contact_export"
    }
)
```

### TypeScript: Adding Custom Variables

All TypeScript builders accept `customVars` in options:

```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";
import { createContactExportDAG } from "./src/builders/ContactExportDAGBuilder";

// Contact Import
const dag = createContactImportDAG("my-bucket", "uploads/contacts.csv", {
  customVars: {
    user_id: "user-123",
    campaign_id: "campaign-456",
    source: "manual_upload",
    priority: "high"
  }
});

// Email Export
const dag = createEmailExportDAG("my-bucket", "exports/contacts.csv", {
  customVars: {
    user_id: "user-123",
    campaign_id: "campaign-456",
    export_type: "email_enrichment"
  }
});

// Contact Export
const dag = createContactExportDAG(
  "my-bucket",
  ["first_name", "last_name", "email"],
  {
    customVars: {
      user_id: "user-123",
      campaign_id: "campaign-456",
      export_type: "contact_export"
    }
  }
);
```

### Using BaseDAGBuilder.addCustomVars()

**Python:**
```python
from dag_builders.base import BaseDAGBuilder

job_data = {
    "s3_bucket": "my-bucket",
    "s3_key": "uploads/contacts.csv"
}

# Add custom variables
job_data = BaseDAGBuilder.add_custom_vars(
    job_data,
    custom_vars={
        "user_id": "user-123",
        "campaign_id": "campaign-456"
    }
)

# job_data now has custom_vars merged in
# {
#     "s3_bucket": "my-bucket",
#     "s3_key": "uploads/contacts.csv",
#     "custom_vars": {
#         "user_id": "user-123",
#         "campaign_id": "campaign-456"
#     }
# }
```

**TypeScript:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

const jobData = {
  s3_bucket: "my-bucket",
  s3_key: "uploads/contacts.csv"
};

// Add custom variables
BaseDAGBuilder.addCustomVars(jobData, {
  user_id: "user-123",
  campaign_id: "campaign-456"
});

// jobData now has custom_vars merged in
// {
//     s3_bucket: "my-bucket",
//     s3_key: "uploads/contacts.csv",
//     custom_vars: {
//         user_id: "user-123",
//         campaign_id: "campaign-456"
//     }
// }
```

---

## Dynamic DAG Patterns

### Pattern 1: Sequential Pipeline with Shared Variables

Create a pipeline where all jobs share common custom variables.

**Python:**
```python
from dag_builders.base import BaseDAGBuilder

def create_shared_variables_pipeline(workflow_id: str, common_vars: dict):
    """Create pipeline with shared custom variables"""
    
    dag = [
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-extract",
            "Extract Data",
            "etl",
            {
                "source": "s3://bucket/data.csv",
                "custom_vars": common_vars  # Shared variables
            },
            edges=[f"{workflow_id}-transform"]
        ),
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-transform",
            "Transform Data",
            "etl",
            {
                "transformations": ["clean", "normalize"],
                "custom_vars": common_vars  # Shared variables
            },
            edges=[f"{workflow_id}-load"]
        ),
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-load",
            "Load to Warehouse",
            "etl",
            {
                "destination": "warehouse",
                "custom_vars": common_vars  # Shared variables
            },
            edges=[]
        )
    ]
    
    return dag

# Usage
workflow_id = BaseDAGBuilder.generate_workflow_id("pipeline")
common_vars = {
    "campaign_id": "campaign-456",
    "date": "2024-01-15",
    "region": "us-east-1"
}

dag = create_shared_variables_pipeline(workflow_id, common_vars)
```

**TypeScript:**
```typescript
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

function createSharedVariablesPipeline(workflowId: string, commonVars: Record<string, unknown>) {
  const dag = [
    BaseDAGBuilder.buildJobNode(
      `${workflowId}-extract`,
      "Extract Data",
      "etl",
      {
        source: "s3://bucket/data.csv",
        custom_vars: commonVars  // Shared variables
      },
      { edges: [`${workflowId}-transform`] }
    ),
    BaseDAGBuilder.buildJobNode(
      `${workflowId}-transform`,
      "Transform Data",
      "etl",
      {
        transformations: ["clean", "normalize"],
        custom_vars: commonVars  // Shared variables
      },
      { edges: [`${workflowId}-load`] }
    ),
    BaseDAGBuilder.buildJobNode(
      `${workflowId}-load`,
      "Load to Warehouse",
      "etl",
      {
        destination: "warehouse",
        custom_vars: commonVars  // Shared variables
      },
      { edges: [] }
    )
  ];
  
  return dag;
}

// Usage
const workflowId = BaseDAGBuilder.generateWorkflowId("pipeline");
const commonVars = {
  campaign_id: "campaign-456",
  date: "2024-01-15",
  region: "us-east-1"
};

const dag = createSharedVariablesPipeline(workflowId, commonVars);
```

### Pattern 2: Conditional DAG Generation

Generate different DAG structures based on conditions.

**Python:**
```python
from dag_builders.contact_import import create_contact_import_dag
from dag_builders.email_export import create_email_export_dag
from dag_builders.base import BaseDAGBuilder

def create_conditional_workflow(file_type: str, s3_key: str, config: dict):
    """Create workflow based on file type"""
    
    workflow_id = BaseDAGBuilder.generate_workflow_id(f"{file_type}-import")
    common_vars = {
        "file_type": file_type,
        "user_id": config.get("user_id"),
        "date": config.get("date")
    }
    
    if file_type == "contacts":
        # Simple import
        dag = create_contact_import_dag(
            s3_bucket=config["bucket"],
            s3_key=s3_key,
            custom_vars=common_vars
        )
    elif file_type == "contacts_with_email":
        # Import + Email enrichment pipeline
        import_dag = create_contact_import_dag(
            s3_bucket=config["bucket"],
            s3_key=s3_key,
            custom_vars=common_vars
        )
        
        import_job_id = import_dag[0]["uuid"]
        
        # Email export depends on import
        export_dag = create_email_export_dag(
            s3_bucket=config["bucket"],
            s3_key=f"exports/{s3_key.split('/')[-1]}",
            custom_vars=common_vars
        )
        
        # Add dependency
        export_dag[0]["edges"] = [import_job_id]
        
        dag = import_dag + export_dag
    else:
        raise ValueError(f"Unknown file type: {file_type}")
    
    return dag

# Usage
config = {
    "bucket": "my-bucket",
    "user_id": "user-123",
    "date": "2024-01-15"
}

# Simple import
dag1 = create_conditional_workflow("contacts", "uploads/contacts.csv", config)

# Import with email enrichment
dag2 = create_conditional_workflow("contacts_with_email", "uploads/contacts.csv", config)
```

**TypeScript:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";
import { createEmailExportDAG } from "./src/builders/EmailExportDAGBuilder";
import { BaseDAGBuilder } from "./src/builders/BaseDAGBuilder";

interface Config {
  bucket: string;
  userId?: string;
  date?: string;
}

function createConditionalWorkflow(
  fileType: string,
  s3Key: string,
  config: Config
) {
  const workflowId = BaseDAGBuilder.generateWorkflowId(`${fileType}-import`);
  const commonVars = {
    file_type: fileType,
    user_id: config.userId,
    date: config.date
  };
  
  if (fileType === "contacts") {
    // Simple import
    return createContactImportDAG(config.bucket, s3Key, {
      customVars: commonVars
    });
  } else if (fileType === "contacts_with_email") {
    // Import + Email enrichment pipeline
    const importDAG = createContactImportDAG(config.bucket, s3Key, {
      customVars: commonVars
    });
    
    const importJobId = importDAG[0].uuid;
    
    // Email export depends on import
    const exportDAG = createEmailExportDAG(config.bucket, `exports/${s3Key.split("/").pop()}`, {
      customVars: commonVars
    });
    
    // Add dependency
    exportDAG[0].edges = [importJobId];
    
    return [...importDAG, ...exportDAG];
  } else {
    throw new Error(`Unknown file type: ${fileType}`);
  }
}

// Usage
const config: Config = {
  bucket: "my-bucket",
  userId: "user-123",
  date: "2024-01-15"
};

// Simple import
const dag1 = createConditionalWorkflow("contacts", "uploads/contacts.csv", config);

// Import with email enrichment
const dag2 = createConditionalWorkflow("contacts_with_email", "uploads/contacts.csv", config);
```

---

## Template-Based DAG Creation

### Using DAGTemplate (Python)

The `DAGTemplate` class provides template-based DAG creation with variable substitution.

**Python Example:**
```python
from utils.templates import DAGTemplate

# Define template
template = {
    "nodes": [
        {
            "uuid": "extract-{date}",
            "job_title": "Extract Data - {date}",
            "job_type": "etl",
            "data": {
                "source": "s3://{bucket}/data/{date}/input.csv",
                "custom_vars": {
                    "date": "{date}",
                    "region": "{region}"
                }
            },
            "edges": ["transform-{date}"]
        },
        {
            "uuid": "transform-{date}",
            "job_title": "Transform Data - {date}",
            "job_type": "etl",
            "data": {
                "transformations": ["clean", "normalize"],
                "custom_vars": {
                    "date": "{date}",
                    "region": "{region}"
                }
            },
            "edges": []
        }
    ]
}

# Process template with variables
variables = {
    "date": "2024-01-15",
    "bucket": "my-bucket",
    "region": "us-east-1"
}

dag = DAGTemplate.process_template(template, variables)

# Result:
# [
#     {
#         "uuid": "extract-2024-01-15",
#         "job_title": "Extract Data - 2024-01-15",
#         "data": {
#             "source": "s3://my-bucket/data/2024-01-15/input.csv",
#             "custom_vars": {
#                 "date": "2024-01-15",
#                 "region": "us-east-1"
#             }
#         },
#         ...
#     },
#     ...
# ]
```

### Extracting Template Variables

```python
from utils.templates import DAGTemplate

template = {
    "nodes": [
        {
            "uuid": "extract-{date}",
            "data": {
                "source": "s3://{bucket}/data/{date}/input.csv",
                "custom_vars": {"region": "{region}"}
            }
        }
    ]
}

# Extract all variable names
variables = DAGTemplate.get_template_variables(template)
# Result: ["date", "bucket", "region"]
```

---

## Querying Jobs by Custom Variables

### PostgreSQL JSONB Querying

Custom variables are stored as JSONB in PostgreSQL, allowing powerful queries:

```sql
-- Find all jobs with specific user_id
SELECT * FROM jobs
WHERE data->'custom_vars'->>'user_id' = 'user-123';

-- Find all jobs with specific campaign_id
SELECT * FROM jobs
WHERE data->'custom_vars'->>'campaign_id' = 'campaign-456';

-- Find all jobs with specific date
SELECT * FROM jobs
WHERE data->'custom_vars'->>'date' = '2024-01-15';

-- Find all jobs with multiple custom variables
SELECT * FROM jobs
WHERE data->'custom_vars'->>'user_id' = 'user-123'
  AND data->'custom_vars'->>'campaign_id' = 'campaign-456';

-- Find all jobs with custom variable containing value (array)
SELECT * FROM jobs
WHERE data->'custom_vars'->'tags' @> '"urgent"'::jsonb;

-- Count jobs by custom variable
SELECT 
    data->'custom_vars'->>'campaign_id' as campaign_id,
    COUNT(*) as job_count
FROM jobs
WHERE data->'custom_vars'->>'campaign_id' IS NOT NULL
GROUP BY data->'custom_vars'->>'campaign_id';
```

### Programmatic Querying (Python)

```python
import requests

def get_jobs_by_custom_var(api_url: str, api_key: str, var_name: str, var_value: str):
    """Query jobs by custom variable using PostgreSQL JSONB query"""
    
    # Note: This requires API support or direct database access
    # For now, fetch all jobs and filter client-side
    
    response = requests.get(
        f"{api_url}/jobs/",
        headers={"X-API-Key": api_key},
        params={"status": ["open", "processing", "completed", "failed"]}
    )
    
    jobs = response.json()["data"]
    
    # Filter by custom variable
    filtered_jobs = [
        job for job in jobs
        if job.get("data", {}).get("custom_vars", {}).get(var_name) == var_value
    ]
    
    return filtered_jobs

# Usage
jobs = get_jobs_by_custom_var(
    "http://localhost:8000",
    "your-api-key",
    "user_id",
    "user-123"
)

for job in jobs:
    print(f"Job: {job['uuid']}, Status: {job['status']}")
```

### Programmatic Querying (TypeScript)

```typescript
async function getJobsByCustomVar(
  apiUrl: string,
  apiKey: string,
  varName: string,
  varValue: string
) {
  // Note: This requires API support or direct database access
  // For now, fetch all jobs and filter client-side
  
  const response = await fetch(
    `${apiUrl}/jobs/?status=open&status=processing&status=completed&status=failed`,
    {
      headers: { "X-API-Key": apiKey }
    }
  );
  
  const result = await response.json();
  const jobs = result.data || [];
  
  // Filter by custom variable
  const filteredJobs = jobs.filter(
    (job: any) =>
      job.data?.custom_vars?.[varName] === varValue
  );
  
  return filteredJobs;
}

// Usage
const jobs = await getJobsByCustomVar(
  "http://localhost:8000",
  "your-api-key",
  "user_id",
  "user-123"
);

for (const job of jobs) {
  console.log(`Job: ${job.uuid}, Status: ${job.status}`);
}
```

---

## Complete Examples

### Example 1: Batch Processing with Custom Variables

**Python:**
```python
import requests
from dag_builders.contact_import import create_contact_import_dag
from datetime import datetime

def create_batch_import_dags(file_list: list, user_id: str, campaign_id: str):
    """Create batch import DAGs with shared custom variables"""
    
    dags = []
    base_date = datetime.now().isoformat()
    
    for idx, file_path in enumerate(file_list):
        dag = create_contact_import_dag(
            s3_bucket="my-uploads-bucket",
            s3_key=file_path,
            custom_vars={
                "user_id": user_id,
                "campaign_id": campaign_id,
                "batch_id": f"batch-{idx+1:03d}",
                "batch_date": base_date,
                "file_path": file_path,
                "source": "batch_upload",
                "upload_type": "bulk"
            }
        )
        dags.extend(dag)
    
    return dags

# Usage
file_list = [
    "uploads/contacts_batch_1.csv",
    "uploads/contacts_batch_2.csv",
    "uploads/contacts_batch_3.csv",
    "uploads/contacts_batch_4.csv",
    "uploads/contacts_batch_5.csv",
]

dags = create_batch_import_dags(
    file_list=file_list,
    user_id="user-123",
    campaign_id="campaign-456"
)

# Submit all DAGs at once
response = requests.post(
    "http://localhost:8000/jobs/bulk-insert/complete-graph",
    headers={"X-API-Key": "your-api-key"},
    json=dags
)

print(f"Submitted {len(dags)} jobs")
```

**TypeScript:**
```typescript
import { createContactImportDAG } from "./src/builders/ContactImportDAGBuilder";

function createBatchImportDAGs(
  fileList: string[],
  userId: string,
  campaignId: string
) {
  const dags: any[] = [];
  const baseDate = new Date().toISOString();
  
  fileList.forEach((filePath, idx) => {
    const dag = createContactImportDAG("my-uploads-bucket", filePath, {
      customVars: {
        user_id: userId,
        campaign_id: campaignId,
        batch_id: `batch-${(idx + 1).toString().padStart(3, "0")}`,
        batch_date: baseDate,
        file_path: filePath,
        source: "batch_upload",
        upload_type: "bulk"
      }
    });
    dags.push(...dag);
  });
  
  return dags;
}

// Usage
const fileList = [
  "uploads/contacts_batch_1.csv",
  "uploads/contacts_batch_2.csv",
  "uploads/contacts_batch_3.csv",
  "uploads/contacts_batch_4.csv",
  "uploads/contacts_batch_5.csv",
];

const dags = createBatchImportDAGs(fileList, "user-123", "campaign-456");

// Submit all DAGs at once
const response = await fetch(
  "http://localhost:8000/jobs/bulk-insert/complete-graph",
  {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-API-Key": "your-api-key"
    },
    body: JSON.stringify(dags)
  }
);

console.log(`Submitted ${dags.length} jobs`);
```

### Example 2: Multi-Step Pipeline with Per-Job Variables

**Python:**
```python
from dag_builders.base import BaseDAGBuilder

def create_multi_step_pipeline(config: dict):
    """Create multi-step pipeline with job-specific custom variables"""
    
    workflow_id = BaseDAGBuilder.generate_workflow_id("pipeline")
    base_vars = {
        "workflow_id": workflow_id,
        "campaign_id": config["campaign_id"],
        "date": config["date"]
    }
    
    dag = [
        # Step 1: Extract
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-extract",
            "Extract Data",
            "etl",
            {
                "source": config["source"],
                "custom_vars": {
                    **base_vars,
                    "step": "extract",
                    "step_number": 1,
                    "total_steps": 3
                }
            },
            edges=[f"{workflow_id}-transform"]
        ),
        
        # Step 2: Transform
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-transform",
            "Transform Data",
            "etl",
            {
                "transformations": config["transformations"],
                "custom_vars": {
                    **base_vars,
                    "step": "transform",
                    "step_number": 2,
                    "total_steps": 3
                }
            },
            edges=[f"{workflow_id}-load"]
        ),
        
        # Step 3: Load
        BaseDAGBuilder.build_job_node(
            f"{workflow_id}-load",
            "Load to Warehouse",
            "etl",
            {
                "destination": config["destination"],
                "custom_vars": {
                    **base_vars,
                    "step": "load",
                    "step_number": 3,
                    "total_steps": 3
                }
            },
            edges=[]
        )
    ]
    
    return dag

# Usage
config = {
    "campaign_id": "campaign-456",
    "date": "2024-01-15",
    "source": "s3://bucket/data.csv",
    "transformations": ["clean", "normalize"],
    "destination": "warehouse"
}

dag = create_multi_step_pipeline(config)
```

---

## Best Practices

### 1. Use Consistent Variable Names

Use consistent naming conventions for custom variables:

```python
# Good: Consistent naming
custom_vars = {
    "user_id": "user-123",
    "campaign_id": "campaign-456",
    "upload_date": "2024-01-15"
}

# Bad: Inconsistent naming
custom_vars = {
    "userId": "user-123",  # camelCase
    "campaign-id": "campaign-456",  # kebab-case
    "uploadDate": "2024-01-15"  # Mixed
}
```

### 2. Group Related Variables

Group related custom variables logically:

```python
# Good: Grouped by purpose
custom_vars = {
    # User context
    "user_id": "user-123",
    "org_id": "org-001",
    
    # Campaign context
    "campaign_id": "campaign-456",
    "campaign_type": "email",
    
    # Temporal context
    "date": "2024-01-15",
    "created_at": "2024-01-15T10:00:00Z"
}

# Bad: No grouping
custom_vars = {
    "user_id": "user-123",
    "date": "2024-01-15",
    "campaign_id": "campaign-456",
    "org_id": "org-001"
}
```

### 3. Use Variables for Filtering/Querying

Store variables that you'll need to query later:

```python
# Good: Variables useful for querying
custom_vars = {
    "user_id": "user-123",  # Query: All jobs for user
    "campaign_id": "campaign-456",  # Query: All jobs for campaign
    "date": "2024-01-15",  # Query: All jobs for date
    "status": "active"  # Query: All active jobs
}

# Bad: Variables not useful for querying
custom_vars = {
    "internal_note": "This is just a note",  # Not queryable
    "debug_info": {"level": "debug"}  # Not easily queryable
}
```

### 4. Limit Variable Size

Keep custom variables small (JSONB field size limits):

```python
# Good: Small, essential variables
custom_vars = {
    "user_id": "user-123",
    "campaign_id": "campaign-456",
    "date": "2024-01-15"
}

# Bad: Too large
custom_vars = {
    "large_data": "x" * 100000,  # Too large
    "full_payload": {...}  # Store in separate field
}
```

### 5. Validate Variables Before Submission

Validate custom variables before creating DAGs:

```python
from dag_builders.base import BaseDAGBuilder

def validate_custom_vars(custom_vars: dict) -> bool:
    """Validate custom variables"""
    # Check required fields
    required = ["user_id", "campaign_id"]
    for field in required:
        if field not in custom_vars:
            return False
    
    # Check types
    if not isinstance(custom_vars["user_id"], str):
        return False
    
    return True

# Usage
custom_vars = {
    "user_id": "user-123",
    "campaign_id": "campaign-456"
}

if not validate_custom_vars(custom_vars):
    raise ValueError("Invalid custom variables")

dag = create_contact_import_dag(
    "my-bucket",
    "uploads/contacts.csv",
    custom_vars=custom_vars
)
```

---

## Task Breakdown

### Phase 1: Understanding Dynamic DAGs ✅

- [x] Understand what dynamic DAGs are
- [x] Learn difference between static and dynamic DAGs
- [x] Study use cases for dynamic DAGs

### Phase 2: Custom Variables Overview ✅

- [x] Understand custom variables structure
- [x] Learn where custom variables are stored (data.custom_vars)
- [x] Study benefits of custom variables (querying, grouping, tracking)

### Phase 3: Dynamic DAG Creation Patterns ✅

- [x] Learn loop-based generation pattern
- [x] Learn date-based dynamic DAGs
- [x] Learn config-driven dynamic DAGs
- [x] Learn multi-tenant dynamic DAGs

### Phase 4: Using Custom Variables in Builders ✅

- [x] Study Python builder custom_vars parameter
- [x] Study TypeScript builder customVars parameter
- [x] Learn BaseDAGBuilder.addCustomVars() method

### Phase 5: Dynamic DAG Patterns ✅

- [x] Learn sequential pipeline with shared variables
- [x] Learn conditional DAG generation
- [x] Study multi-step pipeline patterns

### Phase 6: Template-Based DAG Creation ✅

- [x] Understand DAGTemplate class
- [x] Learn variable substitution in templates
- [x] Study template variable extraction

### Phase 7: Querying Jobs by Custom Variables ✅

- [x] Learn PostgreSQL JSONB querying
- [x] Study programmatic querying (Python/TypeScript)
- [x] Practice filtering jobs by custom variables

### Phase 8: Complete Examples ✅

- [x] Study batch processing examples
- [x] Study multi-step pipeline examples
- [x] Practice creating dynamic DAGs

### Phase 9: Best Practices ✅

- [x] Learn consistent variable naming
- [x] Understand variable grouping
- [x] Study variable validation
- [x] Learn variable size limits

---

## Summary

Creating dynamic DAGs with custom variables enables:

1. **Flexibility**: Generate DAGs based on runtime data
2. **Scalability**: Process multiple files/jobs programmatically
3. **Tracking**: Add metadata for querying and monitoring
4. **Organization**: Group related jobs by custom variables

Key takeaways:
- Use `custom_vars` in builders to add metadata
- Create dynamic DAGs using loops, conditions, and templates
- Query jobs by custom variables using PostgreSQL JSONB
- Follow best practices for variable naming and grouping

---

## Additional Resources

- [HOW_TO_CREATE_DAGS.md](./HOW_TO_CREATE_DAGS.md) - Basic DAG creation guide
- [DAG_BUILDERS_GUIDE.md](./DAG_BUILDERS_GUIDE.md) - Detailed builder documentation
- [Python Examples](../client-libs/python/examples/) - Working examples
- [TypeScript Examples](../client-libs/typescript/examples/) - Working examples
