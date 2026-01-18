/** DAG Validation Utilities

Provides validation utilities for DAG structures before submission to the API.
*/

import { BaseDAGBuilder } from "../builders/BaseDAGBuilder";
import type { DAGDefinition, DAGValidationResult } from "../types/dag.types";

export class DAGValidator {
  /**
   * Validate DAG structure
   */
  static validateDAG(dag: DAGDefinition): DAGValidationResult {
    return BaseDAGBuilder.validateDAG(dag);
  }

  /**
   * Validate all UUIDs in DAG are unique
   */
  static validateUUIDUniqueness(dag: DAGDefinition): DAGValidationResult {
    const uuids = dag.map((node) => node.uuid);
    const uniqueUuids = new Set(uuids);

    if (uuids.length !== uniqueUuids.size) {
      const duplicates = uuids.filter((uuid, index) => uuids.indexOf(uuid) !== index);
      return {
        isValid: false,
        error: `Duplicate UUIDs found: ${[...new Set(duplicates)].join(", ")}`,
      };
    }

    return { isValid: true };
  }

  /**
   * Validate all edge targets exist in DAG
   */
  static validateEdges(dag: DAGDefinition): DAGValidationResult {
    const allUuids = new Set(dag.map((node) => node.uuid));

    for (const node of dag) {
      for (const edgeTarget of node.edges || []) {
        if (!allUuids.has(edgeTarget)) {
          return {
            isValid: false,
            error: `Edge target '${edgeTarget}' from node '${node.uuid}' not found in DAG`,
          };
        }
      }
    }

    return { isValid: true };
  }

  /**
   * Validate all required fields are present
   */
  static validateRequiredFields(dag: DAGDefinition): DAGValidationResult {
    const requiredFields: (keyof typeof dag[0])[] = ["uuid", "job_title", "job_type", "data"];

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

  /**
   * Validate node count is within limits
   */
  static validateNodeCount(dag: DAGDefinition): DAGValidationResult {
    if (!dag || dag.length === 0) {
      return { isValid: false, error: "DAG cannot be empty" };
    }

    if (dag.length > BaseDAGBuilder.MAX_NODES_PER_REQUEST) {
      return {
        isValid: false,
        error: `DAG exceeds maximum nodes limit (${BaseDAGBuilder.MAX_NODES_PER_REQUEST}): ${dag.length} nodes`,
      };
    }

    return { isValid: true };
  }
}
