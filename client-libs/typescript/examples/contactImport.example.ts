/** Contact Import DAG Example

Example demonstrating how to create contact import DAGs.
*/

import { createContactImportDAG } from "../src/builders/ContactImportDAGBuilder";
import { ImportExportDAGBuilder } from "../src/builders/UnifiedDAGBuilder";

// Example 1: Basic contact import
export function exampleBasicContactImport() {
  const dag = createContactImportDAG("my-uploads-bucket", "uploads/contacts_2024.csv");

  console.log("Contact Import DAG:", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 2: Contact import with custom retry settings
export function exampleCustomRetry() {
  const dag = createContactImportDAG("my-uploads-bucket", "uploads/large_contacts_2024.csv", {
    retryCount: 5,
    retryInterval: 10, // 10 minutes between retries
    customVars: {
      source: "manual_upload",
      user_id: "user-123",
    },
  });

  console.log("Contact Import DAG (Custom Retry):", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 3: Using unified builder
export function exampleUnifiedBuilder() {
  const dag = ImportExportDAGBuilder.createImportDAG(
    "contact",
    "my-uploads-bucket",
    "uploads/contacts_2024.csv",
    {
      workflowId: "my-custom-workflow-id",
    }
  );

  console.log("Contact Import DAG (Unified Builder):", JSON.stringify(dag, null, 2));
  return dag;
}

if (require.main === module) {
  console.log("=".repeat(60));
  console.log("Example 1: Basic Contact Import");
  console.log("=".repeat(60));
  exampleBasicContactImport();

  console.log("\n" + "=".repeat(60));
  console.log("Example 2: Custom Retry Settings");
  console.log("=".repeat(60));
  exampleCustomRetry();

  console.log("\n" + "=".repeat(60));
  console.log("Example 3: Unified Builder");
  console.log("=".repeat(60));
  exampleUnifiedBuilder();

  console.log("\n" + "=".repeat(60));
  console.log("Examples complete!");
  console.log("=".repeat(60));
}
