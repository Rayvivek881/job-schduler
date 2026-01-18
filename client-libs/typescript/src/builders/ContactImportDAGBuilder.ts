/** Contact Import DAG Builder

Creates DAGs for importing contacts from CSV files stored in S3.
*/

import { BaseDAGBuilder } from "./BaseDAGBuilder";
import type { DAGDefinition, InsertFileJobData } from "../types/dag.types";
import type { BaseDAGOptions } from "../types/dag.types";

export interface ContactImportDAGOptions extends BaseDAGOptions {}

/**
 * Create DAG for importing contacts from CSV file
 */
export function createContactImportDAG(
  s3Bucket: string,
  s3Key: string,
  options: ContactImportDAGOptions = {}
): DAGDefinition {
  const workflowId =
    options.workflowId || BaseDAGBuilder.generateWorkflowId("contact-import");

  // Build job data matching InsertFileJobData structure
  const jobData: InsertFileJobData = {
    s3_key: s3Key,
    s3_bucket: s3Bucket,
    workflow_uuid: workflowId,
    service: "contact", // Metadata indicating contact service
    workflow_type: "contact_import",
    created_at: new Date().toISOString(),
  };

  // Add custom variables
  const customVars = {
    source: "csv_import",
    destination: "contacts_db",
    ...options.customVars,
  };
  jobData.custom_vars = customVars;

  // Build job node
  return [
    BaseDAGBuilder.buildJobNode(
      workflowId,
      `Import Contacts from CSV: ${s3Key}`,
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
