"""Company Import DAG Builder

Creates DAGs for importing companies from CSV files stored in S3.
"""

from datetime import datetime
from typing import Any, Dict, List, Optional

from .base import BaseDAGBuilder


def create_company_import_dag(
    s3_bucket: str,
    s3_key: str,
    workflow_id: Optional[str] = None,
    retry_count: int = 3,
    retry_interval: int = 5,
    custom_vars: Optional[Dict[str, Any]] = None,
) -> List[Dict[str, Any]]:
    """Create DAG for importing companies from CSV file
    
    This DAG imports company data from a CSV file in S3 into the database.
    The CSV file is processed using the Connectra service's insert_csv_file job type.
    
    Args:
        s3_bucket: S3 bucket containing CSV file
        s3_key: S3 key (path) of CSV file
        workflow_id: Optional workflow UUID (auto-generated if None)
        retry_count: Number of retry attempts
        retry_interval: Minutes between retries
        custom_vars: Additional custom variables to include in data
        
    Returns:
        DAG definition (single job node) ready for API submission
        
    Example:
        >>> dag = create_company_import_dag(
        ...     s3_bucket="my-bucket",
        ...     s3_key="uploads/companies_2024.csv"
        ... )
    """
    if not workflow_id:
        workflow_id = BaseDAGBuilder.generate_workflow_id("company-import")

    # Build job data matching InsertFileJobData structure
    job_data: Dict[str, Any] = {
        "s3_key": s3_key,
        "s3_bucket": s3_bucket,
        "workflow_uuid": workflow_id,
        "service": "company",  # Metadata indicating company service
        "workflow_type": "company_import",
        "created_at": datetime.now().isoformat(),
    }

    # Add custom variables
    custom_vars_dict = {
        "source": "csv_import",
        "destination": "companies_db",
    }
    if custom_vars:
        custom_vars_dict.update(custom_vars)
    job_data["custom_vars"] = custom_vars_dict

    # Build job node
    dag = [
        BaseDAGBuilder.build_job_node(
            uuid=workflow_id,
            job_title=f"Import Companies from CSV: {s3_key}",
            job_type="insert_csv_file",  # Connectra job type
            data=job_data,
            edges=[],
            retry_count=retry_count,
            retry_interval=retry_interval,
        )
    ]

    return dag
