"""Email Export DAG Builder

Creates DAGs for exporting CSV files from S3, enriching with email finder API,
and writing enriched CSV back to S3.
"""

from datetime import datetime
from typing import Any, Dict, List, Optional

from .base import BaseDAGBuilder


def create_email_export_dag(
    s3_bucket: str,
    s3_key: str,
    workflow_id: Optional[str] = None,
    output_s3_key: Optional[str] = None,
    email_api_url: Optional[str] = None,
    email_api_key: Optional[str] = None,
    batch_size: int = 100,
    retry_count: int = 3,
    retry_interval: int = 5,
    custom_vars: Optional[Dict[str, Any]] = None,
) -> List[Dict[str, Any]]:
    """Create DAG for email export enrichment workflow
    
    This DAG reads a CSV from S3, enriches it with email data via email finder API,
    and writes the enriched CSV back to S3.
    
    Args:
        s3_bucket: S3 bucket containing input CSV file
        s3_key: S3 key (path) of input CSV file
        workflow_id: Optional workflow UUID (auto-generated if None)
        output_s3_key: Optional output S3 key (auto-generated if None)
        email_api_url: Email finder API endpoint URL
        email_api_key: Email finder API key
        batch_size: Number of records to process per batch
        retry_count: Number of retry attempts
        retry_interval: Minutes between retries
        custom_vars: Additional custom variables to include in data
        
    Returns:
        DAG definition (single job node) ready for API submission
        
    Example:
        >>> dag = create_email_export_dag(
        ...     s3_bucket="my-bucket",
        ...     s3_key="exports/contacts.csv",
        ...     email_api_url="https://api.emailfinder.com/v1",
        ...     email_api_key="api-key-123"
        ... )
        >>> # Submit to POST /jobs/bulk-insert/complete-graph
    """
    if not workflow_id:
        workflow_id = BaseDAGBuilder.generate_workflow_id("email-export")

    if not output_s3_key:
        # Generate output key: same directory, with timestamp suffix
        if "/" in s3_key:
            base_path, filename = s3_key.rsplit("/", 1)
            name, ext = filename.rsplit(".", 1) if "." in filename else (filename, "csv")
            timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
            output_s3_key = f"{base_path}/{name}_enriched_{timestamp}.{ext}"
        else:
            timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
            output_s3_key = f"{s3_key.rsplit('.', 1)[0]}_enriched_{timestamp}.csv"

    # Build job data
    job_data: Dict[str, Any] = {
        "s3_bucket": s3_bucket,
        "s3_key": s3_key,  # Input file
        "output_s3_key": output_s3_key,  # Output file
        "workflow_uuid": workflow_id,
        "workflow_type": "email_export_enrich",
        "created_at": datetime.now().isoformat(),
    }

    # Add email API configuration if provided
    if email_api_url:
        job_data["email_api_url"] = email_api_url
    if email_api_key:
        job_data["email_api_key"] = email_api_key

    # Add batch size
    job_data["batch_size"] = batch_size

    # Add custom variables
    custom_vars_dict = {
        "source": "s3_csv",
        "destination": "s3_csv_enriched",
        "operation": "email_enrichment",
    }
    if custom_vars:
        custom_vars_dict.update(custom_vars)
    job_data["custom_vars"] = custom_vars_dict

    # Build job node
    dag = [
        BaseDAGBuilder.build_job_node(
            uuid=workflow_id,
            job_title=f"Email Export Enrichment: {s3_key}",
            job_type="email_export_enrich",  # New job type for email enrichment
            data=job_data,
            edges=[],
            retry_count=retry_count,
            retry_interval=retry_interval,
        )
    ]

    return dag
