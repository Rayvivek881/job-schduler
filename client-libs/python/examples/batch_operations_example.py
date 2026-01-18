"""Batch Operations DAG Example

Example demonstrating batch import/export operations and VQL filter usage.
"""

import sys
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from dag_builders.contact_export import create_contact_export_dag
from dag_builders.contact_import import create_contact_import_dag
from dag_builders.unified import ImportExportDAGBuilder
from utils.vql_builder import VQLFilterBuilder

# Example 1: Contact export with keyword filters
def example_contact_export_keyword_filter():
    """Create contact export DAG with keyword matching"""
    # Build keyword match filter
    keyword_filter = VQLFilterBuilder.keyword_match(
        must={
            "country": ["united states"],
            "title": ["VP", "Director", "Manager"],
        }
    )

    # Combine filters into where clause
    where = VQLFilterBuilder.combine_filters(keyword_match=keyword_filter)

    # Create export DAG
    dag = create_contact_export_dag(
        s3_bucket="my-exports-bucket",
        select_columns=["first_name", "last_name", "email", "title", "company_name"],
        where=where,
        limit=1000,
        order_by=[{"order_by": "created_at", "order_direction": "desc"}],
    )

    print("Contact Export DAG (Keyword Filter):")
    print(dag)
    return dag


# Example 2: Contact export with text match filter
def example_contact_export_text_match():
    """Create contact export DAG with text matching"""
    # Build text match filters
    text_matches = [
        VQLFilterBuilder.text_match(
            text_value="software engineer",
            filter_key="title",
            search_type="phrase",
            fuzzy=True,
        ),
        VQLFilterBuilder.text_match(
            text_value="technology",
            filter_key="industry",
            search_type="match",
        ),
    ]

    # Combine filters
    where = VQLFilterBuilder.combine_filters(text_matches=text_matches)

    # Create export DAG
    dag = create_contact_export_dag(
        s3_bucket="my-exports-bucket",
        select_columns=["first_name", "last_name", "email", "title", "company_name"],
        where=where,
        limit=500,
    )

    print("Contact Export DAG (Text Match):")
    print(dag)
    return dag


# Example 3: Batch import operations
def example_batch_imports():
    """Create multiple import DAGs for batch processing"""
    import_files = [
        "uploads/contacts_batch_1.csv",
        "uploads/contacts_batch_2.csv",
        "uploads/contacts_batch_3.csv",
    ]

    dags = []
    for s3_key in import_files:
        dag = create_contact_import_dag(
            s3_bucket="my-uploads-bucket",
            s3_key=s3_key,
            custom_vars={"batch_id": f"batch-{s3_key.split('_')[-1].split('.')[0]}"},
        )
        dags.extend(dag)

    print(f"Batch Import DAGs ({len(dags)} jobs):")
    for dag_node in dags:
        print(f"  - {dag_node['job_title']}")

    return dags


# Example 4: Complex VQL query
def example_complex_vql_query():
    """Create export DAG with complex VQL query"""
    # Build complex filter with multiple conditions
    keyword_filter = VQLFilterBuilder.keyword_match(
        must={"country": ["united states"], "seniority": ["senior"]},
        must_not={"title": ["intern", "associate"]},
    )

    text_matches = [
        VQLFilterBuilder.text_match(
            text_value="engineering",
            filter_key="department",
            search_type="match",
        ),
    ]

    where = VQLFilterBuilder.combine_filters(
        keyword_match=keyword_filter,
        text_matches=text_matches,
    )

    # Create export DAG
    dag = ImportExportDAGBuilder.create_export_dag(
        service="contact",
        s3_bucket="my-exports-bucket",
        select_columns=[
            "first_name",
            "last_name",
            "email",
            "title",
            "company_name",
            "country",
        ],
        where=where,
        limit=2000,
        order_by=[
            {"order_by": "created_at", "order_direction": "desc"},
            {"order_by": "email", "order_direction": "asc"},
        ],
    )

    print("Complex Export DAG:")
    print(dag)
    return dag


if __name__ == "__main__":
    print("=" * 60)
    print("Example 1: Contact Export with Keyword Filter")
    print("=" * 60)
    example_contact_export_keyword_filter()

    print("\n" + "=" * 60)
    print("Example 2: Contact Export with Text Match")
    print("=" * 60)
    example_contact_export_text_match()

    print("\n" + "=" * 60)
    print("Example 3: Batch Imports")
    print("=" * 60)
    example_batch_imports()

    print("\n" + "=" * 60)
    print("Example 4: Complex VQL Query")
    print("=" * 60)
    example_complex_vql_query()

    print("\n" + "=" * 60)
    print("Examples complete!")
    print("=" * 60)
