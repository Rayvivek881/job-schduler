/** Email Export DAG Builder

Creates DAGs for exporting CSV files from S3, enriching with email finder API,
and writing enriched CSV back to S3.
*/

import { BaseDAGBuilder } from "./BaseDAGBuilder";
import type { DAGDefinition, EmailExportJobData } from "../types/dag.types";
import type { BaseDAGOptions } from "../types/dag.types";

export interface EmailExportDAGOptions extends BaseDAGOptions {
  outputS3Key?: string;
  emailApiUrl?: string;
  emailApiKey?: string;
  batchSize?: number;
}

/**
 * Create DAG for email export enrichment workflow
 */
export function createEmailExportDAG(
  s3Bucket: string,
  s3Key: string,
  options: EmailExportDAGOptions = {}
): DAGDefinition {
  const workflowId =
    options.workflowId || BaseDAGBuilder.generateWorkflowId("email-export");

  let outputS3Key = options.outputS3Key;
  if (!outputS3Key) {
    // Generate output key: same directory, with timestamp suffix
    if (s3Key.includes("/")) {
      const parts = s3Key.split("/");
      const filename = parts[parts.length - 1];
      const pathPrefix = parts.slice(0, -1).join("/");
      const nameParts = filename.split(".");
      const name = nameParts[0];
      const ext = nameParts.length > 1 ? nameParts[nameParts.length - 1] : "csv";
      const timestamp = new Date().toISOString().replace(/[-:T]/g, "").substring(0, 15);
      outputS3Key = `${pathPrefix}/${name}_enriched_${timestamp}.${ext}`;
    } else {
      const nameParts = s3Key.split(".");
      const name = nameParts[0];
      const timestamp = new Date().toISOString().replace(/[-:T]/g, "").substring(0, 15);
      outputS3Key = `${name}_enriched_${timestamp}.csv`;
    }
  }

  // Build job data
  const jobData: EmailExportJobData = {
    s3_bucket: s3Bucket,
    s3_key: s3Key, // Input file
    output_s3_key: outputS3Key, // Output file
    workflow_uuid: workflowId,
    workflow_type: "email_export_enrich",
    created_at: new Date().toISOString(),
  };

  // Add email API configuration if provided
  if (options.emailApiUrl) {
    jobData.email_api_url = options.emailApiUrl;
  }
  if (options.emailApiKey) {
    jobData.email_api_key = options.emailApiKey;
  }

  // Add batch size
  jobData.batch_size = options.batchSize || 100;

  // Add custom variables
  const customVars = {
    source: "s3_csv",
    destination: "s3_csv_enriched",
    operation: "email_enrichment",
    ...options.customVars,
  };
  jobData.custom_vars = customVars;

  // Build job node
  return [
    BaseDAGBuilder.buildJobNode(
      workflowId,
      `Email Export Enrichment: ${s3Key}`,
      "email_export_enrich", // New job type for email enrichment
      jobData,
      {
        edges: [],
        retryCount: options.retryCount,
        retryInterval: options.retryInterval,
      }
    ),
  ];
}
