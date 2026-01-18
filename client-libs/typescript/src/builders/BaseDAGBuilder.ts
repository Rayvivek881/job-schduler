/** Base DAG Builder Class

Provides common functionality for all DAG builders including UUID generation,
workflow ID creation, and job node construction.
*/

import type { BaseDAGOptions, DAGDefinition, DAGValidationResult, JobData, JobNode } from "../types/dag.types";

export class BaseDAGBuilder {
  static readonly MAX_NODES_PER_REQUEST = 5000; // From constants.MaxNodesPerRequest

  /**
   * Generate a unique UUID string
   * Uses crypto.randomUUID() in Node.js 14.17.0+ or browser with Web Crypto API
   * Falls back to a simple UUID v4 implementation if not available
   */
  static generateUUID(prefix?: string): string {
    let uuid: string;
    
    // Try to use crypto.randomUUID() if available (Node.js 14.17.0+ or browser)
    if (typeof crypto !== "undefined" && crypto.randomUUID) {
      uuid = crypto.randomUUID().replace(/-/g, "");
    } else {
      // Fallback: Simple UUID v4 generation
      uuid = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx".replace(/[x]/g, (c) => {
        const r = (Math.random() * 16) | 0;
        const v = c === "x" ? r : (r & 0x3) | 0x8;
        return v.toString(16);
      });
    }
    
    return prefix ? `${prefix}-${uuid}` : uuid;
  }

  /**
   * Generate a workflow ID with timestamp and optional suffix
   */
  static generateWorkflowId(workflowType: string, suffix?: string): string {
    const timestamp = new Date().toISOString().replace(/[-:T]/g, "").split(".")[0];
    const shortTimestamp = timestamp.replace(/(\d{4})(\d{2})(\d{2})(\d{6})/, "$1$2$3-$4");
    const finalSuffix = suffix || BaseDAGBuilder.generateUUID().substring(0, 8);
    return `${workflowType}-${shortTimestamp}-${finalSuffix}`;
  }

  /**
   * Build a job node dictionary for DAG submission
   */
  static buildJobNode(
    uuid: string,
    jobTitle: string,
    jobType: string,
    data: JobData,
    options: {
      edges?: string[];
      retryCount?: number;
      retryInterval?: number;
      runAfter?: Date;
    } = {}
  ): JobNode {
    const jobNode: JobNode = {
      uuid,
      job_title: jobTitle,
      job_type: jobType,
      data,
      retry_count: options.retryCount ?? 3,
      retry_interval: options.retryInterval ?? 5,
      edges: options.edges || [],
    };

    if (options.runAfter) {
      jobNode.run_after = options.runAfter.toISOString();
    }

    return jobNode;
  }

  /**
   * Add custom variables to job data
   */
  static addCustomVars(data: JobData, customVars?: Record<string, unknown>): JobData {
    if (customVars) {
      if (!data.custom_vars) {
        data.custom_vars = {};
      }
      data.custom_vars = { ...data.custom_vars, ...customVars };
    }
    return data;
  }

  /**
   * Validate DAG structure before submission
   */
  static validateDAG(dag: DAGDefinition): DAGValidationResult {
    if (!dag || dag.length === 0) {
      return { isValid: false, error: "DAG cannot be empty" };
    }

    if (dag.length > BaseDAGBuilder.MAX_NODES_PER_REQUEST) {
      return {
        isValid: false,
        error: `DAG exceeds maximum nodes limit (${BaseDAGBuilder.MAX_NODES_PER_REQUEST})`,
      };
    }

    // Check for duplicate UUIDs
    const uuids = dag.map((node) => node.uuid);
    const uniqueUuids = new Set(uuids);
    if (uuids.length !== uniqueUuids.size) {
      const duplicates = uuids.filter((uuid, index) => uuids.indexOf(uuid) !== index);
      return {
        isValid: false,
        error: `Duplicate UUIDs found: ${[...new Set(duplicates)].join(", ")}`,
      };
    }

    // Check all edge targets exist
    const allUuids = new Set(uuids);
    for (const node of dag) {
      for (const edgeTarget of node.edges || []) {
        if (!allUuids.has(edgeTarget)) {
          return {
            isValid: false,
            error: `Edge target '${edgeTarget}' not found in DAG nodes`,
          };
        }
      }
    }

    // Validate required fields
    const requiredFields: (keyof JobNode)[] = ["uuid", "job_title", "job_type", "data"];
    for (const node of dag) {
      for (const field of requiredFields) {
        if (!(field in node)) {
          return {
            isValid: false,
            error: `Required field '${field}' missing in node ${node.uuid || "unknown"}`,
          };
        }
      }
    }

    return { isValid: true };
  }
}
