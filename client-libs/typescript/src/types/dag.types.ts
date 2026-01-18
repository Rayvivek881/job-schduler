/** Type definitions for DAG (Directed Acyclic Graph) structures */

/**
 * Job node structure for DAG submission
 */
export interface JobNode {
  uuid: string;
  job_title: string;
  job_type: string;
  data: JobData;
  retry_count?: number;
  retry_interval?: number;
  run_after?: string; // ISO 8601 datetime string
  edges: string[]; // Array of target UUIDs this job depends on
}

/**
 * Job data payload (stored as JSONB in database)
 */
export interface JobData {
  [key: string]: unknown;
  custom_vars?: Record<string, unknown>;
  workflow_uuid?: string;
  workflow_type?: string;
  created_at?: string;
}

/**
 * Insert file job data structure
 */
export interface InsertFileJobData extends JobData {
  s3_key: string;
  s3_bucket: string;
  service?: "contact" | "company";
}

/**
 * Export file job data structure
 */
export interface ExportFileJobData extends JobData {
  s3_bucket: string;
  service: "contact" | "company";
  vql: VQLQuery;
}

/**
 * Email export enrichment job data structure
 */
export interface EmailExportJobData extends JobData {
  s3_bucket: string;
  s3_key: string; // Input file
  output_s3_key?: string; // Output file (auto-generated if not provided)
  email_api_url?: string;
  email_api_key?: string;
  batch_size?: number;
}

/**
 * DAG definition (array of job nodes)
 */
export type DAGDefinition = JobNode[];

/**
 * DAG validation result
 */
export interface DAGValidationResult {
  isValid: boolean;
  error?: string;
}

/**
 * Common DAG builder options
 */
export interface BaseDAGOptions {
  workflowId?: string;
  retryCount?: number;
  retryInterval?: number;
  customVars?: Record<string, unknown>;
}
