/** Email Export DAG Example

Example demonstrating how to create an email export enrichment DAG.
*/

import { createEmailExportDAG } from "../src/builders/EmailExportDAGBuilder";

// Example 1: Basic email export
export function exampleBasicEmailExport() {
  const dag = createEmailExportDAG(
    "my-exports-bucket",
    "exports/contacts_2024.csv",
    {
      emailApiUrl: "https://api.emailfinder.com/v1",
      emailApiKey: "api-key-123",
      batchSize: 100,
    }
  );

  console.log("Email Export DAG:", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 2: Email export with custom output path
export function exampleCustomOutputPath() {
  const dag = createEmailExportDAG(
    "my-exports-bucket",
    "exports/contacts_2024.csv",
    {
      outputS3Key: "exports/enriched/contacts_2024_enriched.csv",
      emailApiUrl: "https://api.emailfinder.com/v1",
      emailApiKey: "api-key-123",
      batchSize: 50,
      retryCount: 5,
    }
  );

  console.log("Email Export DAG (Custom Output):", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 3: Email export with custom variables
export function exampleCustomVars() {
  const dag = createEmailExportDAG(
    "my-exports-bucket",
    "exports/contacts_2024.csv",
    {
      emailApiUrl: "https://api.emailfinder.com/v1",
      emailApiKey: "api-key-123",
      customVars: {
        user_id: "user-123",
        campaign_id: "campaign-456",
        priority: "high",
      },
    }
  );

  console.log("Email Export DAG (Custom Vars):", JSON.stringify(dag, null, 2));
  return dag;
}

if (require.main === module) {
  console.log("=".repeat(60));
  console.log("Example 1: Basic Email Export");
  console.log("=".repeat(60));
  exampleBasicEmailExport();

  console.log("\n" + "=".repeat(60));
  console.log("Example 2: Custom Output Path");
  console.log("=".repeat(60));
  exampleCustomOutputPath();

  console.log("\n" + "=".repeat(60));
  console.log("Example 3: Custom Variables");
  console.log("=".repeat(60));
  exampleCustomVars();

  console.log("\n" + "=".repeat(60));
  console.log("Examples complete! Use these DAGs with:");
  console.log("POST /jobs/bulk-insert/complete-graph");
  console.log("=".repeat(60));
}
