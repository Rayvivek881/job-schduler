/** Company Export DAG Builder

Creates DAGs for exporting companies from database to CSV files in S3 using VQL filters.
*/

import { BaseDAGBuilder } from "./BaseDAGBuilder";
import { VQLFilterBuilder } from "../utils/VQLBuilder";
import type { DAGDefinition, ExportFileJobData } from "../types/dag.types";
import type { BaseDAGOptions } from "../types/dag.types";
import type { CompanyConfig, FilterOrder, WhereStruct } from "../types/vql.types";

export interface CompanyExportDAGOptions extends BaseDAGOptions {
  where?: WhereStruct;
  orderBy?: FilterOrder[];
  limit?: number;
  companyConfig?: CompanyConfig;
}

/**
 * Create DAG for exporting companies to CSV file
 */
export function createCompanyExportDAG(
  s3Bucket: string,
  selectColumns: string[],
  options: CompanyExportDAGOptions = {}
): DAGDefinition {
  const workflowId =
    options.workflowId || BaseDAGBuilder.generateWorkflowId("company-export");

  // Ensure order_by includes UUID for stable pagination (if not specified)
  let orderBy = options.orderBy;
  if (!orderBy) {
    orderBy = [{ order_by: "uuid", order_direction: "desc" }];
  } else {
    // Check if uuid is already in order_by
    const hasUuidOrder = orderBy.some((order) => order.order_by === "uuid");
    if (!hasUuidOrder) {
      // Prepend UUID ordering for stable pagination
      orderBy = [{ order_by: "uuid", order_direction: "desc" }, ...orderBy];
    }
  }

  // Build VQL query
  const vqlQuery = VQLFilterBuilder.buildVQLQuery({
    where: options.where,
    selectColumns,
    orderBy,
    limit: options.limit,
    companyConfig: options.companyConfig,
  });

  // Build job data matching ExportFileJobData structure
  const jobData: ExportFileJobData = {
    s3_bucket: s3Bucket,
    service: "company",
    vql: vqlQuery,
    workflow_uuid: workflowId,
    workflow_type: "company_export",
    created_at: new Date().toISOString(),
  };

  // Add custom variables
  const customVars = {
    source: "companies_db",
    destination: "s3_csv",
    ...options.customVars,
  };
  jobData.custom_vars = customVars;

  // Build job node
  return [
    BaseDAGBuilder.buildJobNode(
      workflowId,
      `Export Companies to CSV (limit: ${options.limit || "unlimited"})`,
      "export_csv_file", // Connectra job type
      jobData,
      {
        edges: [],
        retryCount: options.retryCount,
        retryInterval: options.retryInterval,
      }
    ),
  ];
}
