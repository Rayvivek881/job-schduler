/** Unified DAG Builder

Provides a unified interface for creating import/export DAGs for different services.
*/

import { createCompanyExportDAG } from "./CompanyExportDAGBuilder";
import { createCompanyImportDAG } from "./CompanyImportDAGBuilder";
import { createContactExportDAG } from "./ContactExportDAGBuilder";
import { createContactImportDAG } from "./ContactImportDAGBuilder";
import { createEmailExportDAG } from "./EmailExportDAGBuilder";
import type { DAGDefinition } from "../types/dag.types";
import type { BaseDAGOptions } from "../types/dag.types";
import type { CompanyConfig, FilterOrder, WhereStruct } from "../types/vql.types";

export interface ImportDAGOptions extends BaseDAGOptions {}

export interface ExportDAGOptions extends BaseDAGOptions {
  where?: WhereStruct;
  orderBy?: FilterOrder[];
  limit?: number;
  companyConfig?: CompanyConfig;
}

export interface EmailExportDAGOptions extends BaseDAGOptions {
  outputS3Key?: string;
  emailApiUrl?: string;
  emailApiKey?: string;
  batchSize?: number;
}

export class ImportExportDAGBuilder {
  /**
   * Create import DAG for specified service
   */
  static createImportDAG(
    service: "contact" | "company",
    s3Bucket: string,
    s3Key: string,
    options: ImportDAGOptions = {}
  ): DAGDefinition {
    if (service === "contact") {
      return createContactImportDAG(s3Bucket, s3Key, options);
    } else if (service === "company") {
      return createCompanyImportDAG(s3Bucket, s3Key, options);
    } else {
      throw new Error(`Unsupported service: ${service}. Must be 'contact' or 'company'`);
    }
  }

  /**
   * Create export DAG for specified service
   */
  static createExportDAG(
    service: "contact" | "company",
    s3Bucket: string,
    selectColumns: string[],
    options: ExportDAGOptions = {}
  ): DAGDefinition {
    if (service === "contact") {
      return createContactExportDAG(s3Bucket, selectColumns, options);
    } else if (service === "company") {
      return createCompanyExportDAG(s3Bucket, selectColumns, options);
    } else {
      throw new Error(`Unsupported service: ${service}. Must be 'contact' or 'company'`);
    }
  }

  /**
   * Create email export enrichment DAG
   */
  static createEmailExportDAG(
    s3Bucket: string,
    s3Key: string,
    options: EmailExportDAGOptions = {}
  ): DAGDefinition {
    return createEmailExportDAG(s3Bucket, s3Key, options);
  }
}

/**
 * Factory function to create import DAG
 */
export function createImportDAG(
  service: "contact" | "company",
  s3Bucket: string,
  s3Key: string,
  options: ImportDAGOptions = {}
): DAGDefinition {
  return ImportExportDAGBuilder.createImportDAG(service, s3Bucket, s3Key, options);
}

/**
 * Factory function to create export DAG
 */
export function createExportDAG(
  service: "contact" | "company",
  s3Bucket: string,
  selectColumns: string[],
  options: ExportDAGOptions = {}
): DAGDefinition {
  return ImportExportDAGBuilder.createExportDAG(service, s3Bucket, selectColumns, options);
}
