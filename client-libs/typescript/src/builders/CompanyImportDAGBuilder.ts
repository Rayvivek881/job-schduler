/** Company Import DAG Builder

Creates DAGs for importing companies from CSV files stored in S3.
*/

import { BaseDAGBuilder } from "./BaseDAGBuilder";
import type { DAGDefinition, InsertFileJobData } from "../types/dag.types";
import type { BaseDAGOptions } from "../types/dag.types";

export interface CompanyImportDAGOptions extends BaseDAGOptions {}

/**
 * Create DAG for importing companies from CSV file
 */
export function createCompanyImportDAG(
  s3Bucket: string,
  s3Key: string,
  options: CompanyImportDAGOptions = {}
): DAGDefinition {
  const workflowId =
    options.workflowId || BaseDAGBuilder.generateWorkflowId("company-import");

  // Build job data matching InsertFileJobData structure
  const jobData: InsertFileJobData = {
    s3_key: s3Key,
    s3_bucket: s3Bucket,
    workflow_uuid: workflowId,
    service: "company", // Metadata indicating company service
    workflow_type: "company_import",
    created_at: new Date().toISOString(),
  };

  // Add custom variables
  const customVars = {
    source: "csv_import",
    destination: "companies_db",
    ...options.customVars,
  };
  jobData.custom_vars = customVars;

  // Build job node
  return [
    BaseDAGBuilder.buildJobNode(
      workflowId,
      `Import Companies from CSV: ${s3Key}`,
      "insert_csv_file", // Connectra job type
      jobData,
      {
        edges: [],
        retryCount: options.retryCount,
        retryInterval: options.retryInterval,
      }
    ),
  ];
}
