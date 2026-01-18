/** Batch Operations DAG Example

Example demonstrating batch import/export operations and VQL filter usage.
*/

import { createContactExportDAG } from "../src/builders/ContactExportDAGBuilder";
import { createContactImportDAG } from "../src/builders/ContactImportDAGBuilder";
import { ImportExportDAGBuilder } from "../src/builders/UnifiedDAGBuilder";
import { VQLFilterBuilder } from "../src/utils/VQLBuilder";

// Example 1: Contact export with keyword filters
export function exampleContactExportKeywordFilter() {
  // Build keyword match filter
  const keywordFilter = VQLFilterBuilder.keywordMatch({
    must: {
      country: ["united states"],
      title: ["VP", "Director", "Manager"],
    },
  });

  // Combine filters into where clause
  const where = VQLFilterBuilder.combineFilters({ keywordMatch: keywordFilter });

  // Create export DAG
  const dag = createContactExportDAG(
    "my-exports-bucket",
    ["first_name", "last_name", "email", "title", "company_name"],
    {
      where,
      limit: 1000,
      orderBy: [{ order_by: "created_at", order_direction: "desc" }],
    }
  );

  console.log("Contact Export DAG (Keyword Filter):", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 2: Contact export with text match filter
export function exampleContactExportTextMatch() {
  // Build text match filters
  const textMatches = [
    VQLFilterBuilder.textMatch({
      textValue: "software engineer",
      filterKey: "title",
      searchType: "phrase",
      fuzzy: true,
    }),
    VQLFilterBuilder.textMatch({
      textValue: "technology",
      filterKey: "industry",
      searchType: "match",
    }),
  ];

  // Combine filters
  const where = VQLFilterBuilder.combineFilters({ textMatches });

  // Create export DAG
  const dag = createContactExportDAG(
    "my-exports-bucket",
    ["first_name", "last_name", "email", "title", "company_name"],
    {
      where,
      limit: 500,
    }
  );

  console.log("Contact Export DAG (Text Match):", JSON.stringify(dag, null, 2));
  return dag;
}

// Example 3: Batch import operations
export function exampleBatchImports() {
  const importFiles = [
    "uploads/contacts_batch_1.csv",
    "uploads/contacts_batch_2.csv",
    "uploads/contacts_batch_3.csv",
  ];

  const dags: any[] = [];
  for (const s3Key of importFiles) {
    const dag = createContactImportDAG("my-uploads-bucket", s3Key, {
      customVars: {
        batch_id: `batch-${s3Key.split("_").pop()?.split(".")[0]}`,
      },
    });
    dags.push(...dag);
  }

  console.log(`Batch Import DAGs (${dags.length} jobs):`);
  for (const dagNode of dags) {
    console.log(`  - ${dagNode.job_title}`);
  }

  return dags;
}

// Example 4: Complex VQL query
export function exampleComplexVQLQuery() {
  // Build complex filter with multiple conditions
  const keywordFilter = VQLFilterBuilder.keywordMatch({
    must: { country: ["united states"], seniority: ["senior"] },
    mustNot: { title: ["intern", "associate"] },
  });

  const textMatches = [
    VQLFilterBuilder.textMatch({
      textValue: "engineering",
      filterKey: "department",
      searchType: "match",
    }),
  ];

  const where = VQLFilterBuilder.combineFilters({
    keywordMatch: keywordFilter,
    textMatches,
  });

  // Create export DAG
  const dag = ImportExportDAGBuilder.createExportDAG(
    "contact",
    "my-exports-bucket",
    ["first_name", "last_name", "email", "title", "company_name", "country"],
    {
      where,
      limit: 2000,
      orderBy: [
        { order_by: "created_at", order_direction: "desc" },
        { order_by: "email", order_direction: "asc" },
      ],
    }
  );

  console.log("Complex Export DAG:", JSON.stringify(dag, null, 2));
  return dag;
}

if (require.main === module) {
  console.log("=".repeat(60));
  console.log("Example 1: Contact Export with Keyword Filter");
  console.log("=".repeat(60));
  exampleContactExportKeywordFilter();

  console.log("\n" + "=".repeat(60));
  console.log("Example 2: Contact Export with Text Match");
  console.log("=".repeat(60));
  exampleContactExportTextMatch();

  console.log("\n" + "=".repeat(60));
  console.log("Example 3: Batch Imports");
  console.log("=".repeat(60));
  exampleBatchImports();

  console.log("\n" + "=".repeat(60));
  console.log("Example 4: Complex VQL Query");
  console.log("=".repeat(60));
  exampleComplexVQLQuery();

  console.log("\n" + "=".repeat(60));
  console.log("Examples complete!");
  console.log("=".repeat(60));
}
