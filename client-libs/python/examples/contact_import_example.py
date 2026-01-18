"""Contact Import DAG Example

Example demonstrating how to create contact import DAGs.
"""

import sys
from pathlib import Path

# Add parent directory to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent))

from dag_builders.contact_import import create_contact_import_dag

# Example 1: Basic contact import
def example_basic_contact_import():
    """Create a basic contact import DAG"""
    dag = create_contact_import_dag(
        s3_bucket="my-uploads-bucket",
        s3_key="uploads/contacts_2024.csv",
    )

    print("Contact Import DAG:")
    print(dag)
    return dag


# Example 2: Contact import with custom retry settings
def example_custom_retry():
    """Create contact import DAG with custom retry configuration"""
    dag = create_contact_import_dag(
        s3_bucket="my-uploads-bucket",
        s3_key="uploads/large_contacts_2024.csv",
        retry_count=5,
        retry_interval=10,  # 10 minutes between retries
        custom_vars={
            "source": "manual_upload",
            "user_id": "user-123",
        },
    )

    print("Contact Import DAG (Custom Retry):")
    print(dag)
    return dag


# Example 3: Using unified builder
def example_unified_builder():
    """Create contact import using unified builder"""
    from dag_builders.unified import ImportExportDAGBuilder

    dag = ImportExportDAGBuilder.create_import_dag(
        service="contact",
        s3_bucket="my-uploads-bucket",
        s3_key="uploads/contacts_2024.csv",
        workflow_id="my-custom-workflow-id",
    )

    print("Contact Import DAG (Unified Builder):")
    print(dag)
    return dag


if __name__ == "__main__":
    print("=" * 60)
    print("Example 1: Basic Contact Import")
    print("=" * 60)
    example_basic_contact_import()

    print("\n" + "=" * 60)
    print("Example 2: Custom Retry Settings")
    print("=" * 60)
    example_custom_retry()

    print("\n" + "=" * 60)
    print("Example 3: Unified Builder")
    print("=" * 60)
    example_unified_builder()

    print("\n" + "=" * 60)
    print("Examples complete!")
    print("=" * 60)
