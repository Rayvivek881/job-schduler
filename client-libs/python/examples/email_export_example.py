"""Email Export DAG Example

Example demonstrating how to create an email export enrichment DAG.
"""

import sys
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from dag_builders.email_export import create_email_export_dag

# Example 1: Basic email export
def example_basic_email_export():
    """Create a basic email export DAG"""
    dag = create_email_export_dag(
        s3_bucket="my-exports-bucket",
        s3_key="exports/contacts_2024.csv",
        email_api_url="https://api.emailfinder.com/v1",
        email_api_key="api-key-123",
        batch_size=100,
    )

    print("Email Export DAG:")
    print(dag)
    return dag


# Example 2: Email export with custom output path
def example_custom_output_path():
    """Create email export DAG with custom output path"""
    dag = create_email_export_dag(
        s3_bucket="my-exports-bucket",
        s3_key="exports/contacts_2024.csv",
        output_s3_key="exports/enriched/contacts_2024_enriched.csv",
        email_api_url="https://api.emailfinder.com/v1",
        email_api_key="api-key-123",
        batch_size=50,
        retry_count=5,
    )

    print("Email Export DAG (Custom Output):")
    print(dag)
    return dag


# Example 3: Email export with custom variables
def example_custom_vars():
    """Create email export DAG with custom metadata"""
    dag = create_email_export_dag(
        s3_bucket="my-exports-bucket",
        s3_key="exports/contacts_2024.csv",
        email_api_url="https://api.emailfinder.com/v1",
        email_api_key="api-key-123",
        custom_vars={
            "user_id": "user-123",
            "campaign_id": "campaign-456",
            "priority": "high",
        },
    )

    print("Email Export DAG (Custom Vars):")
    print(dag)
    return dag


if __name__ == "__main__":
    print("=" * 60)
    print("Example 1: Basic Email Export")
    print("=" * 60)
    example_basic_email_export()

    print("\n" + "=" * 60)
    print("Example 2: Custom Output Path")
    print("=" * 60)
    example_custom_output_path()

    print("\n" + "=" * 60)
    print("Example 3: Custom Variables")
    print("=" * 60)
    example_custom_vars()

    print("\n" + "=" * 60)
    print("Examples complete! Use these DAGs with:")
    print("POST /jobs/bulk-insert/complete-graph")
    print("=" * 60)
