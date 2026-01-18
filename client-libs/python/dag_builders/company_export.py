"""Company Export DAG Builder

Creates DAGs for exporting companies from database to CSV files in S3 using VQL filters.
"""

from datetime import datetime
from typing import Any, Dict, List, Optional

from .base import BaseDAGBuilder
from ..utils.vql_builder import VQLFilterBuilder


def create_company_export_dag(
    s3_bucket: str,
    select_columns: List[str],
    workflow_id: Optional[str] = None,
    where: Optional[Dict[str, Any]] = None,
    order_by: Optional[List[Dict[str, str]]] = None,
    limit: Optional[int] = None,
    company_config: Optional[Dict[str, Any]] = None,
    retry_count: int = 3,
    retry_interval: int = 5,
    custom_vars: Optional[Dict[str, Any]] = None,
) -> List[Dict[str, Any]]:
    """Create DAG for exporting companies to CSV file
    
    This DAG exports filtered company data from the database to a CSV file in S3.
    The export uses VQL (Vivek Query Language) for filtering and selecting data.
    
    Args:
        s3_bucket: S3 bucket to store exported CSV
        select_columns: List of column names to export (required)
        workflow_id: Optional workflow UUID (auto-generated if None)
        where: VQL where clause with filters (keyword_match, text_matches, range_query)
        order_by: List of ordering rules [{"order_by": "field", "order_direction": "asc/desc"}]
        limit: Maximum number of records to export
        company_config: Company-specific configuration (populate, select_columns)
        retry_count: Number of retry attempts
        retry_interval: Minutes between retries
        custom_vars: Additional custom variables to include in data
        
    Returns:
        DAG definition (single job node) ready for API submission
        
    Example:
        >>> # Export companies with keyword filter
        >>> where = VQLFilterBuilder.combine_filters(
        ...     keyword_match=VQLFilterBuilder.keyword_match(
        ...         must={"industry": ["technology"], "country": ["united states"]}
        ...     )
        ... )
        >>> dag = create_company_export_dag(
        ...     s3_bucket="my-bucket",
        ...     select_columns=["name", "industry", "website", "employee_count"],
        ...     where=where,
        ...     limit=500
        ... )
    """
    if not workflow_id:
        workflow_id = BaseDAGBuilder.generate_workflow_id("company-export")

    # Ensure order_by includes UUID for stable pagination (if not specified)
    if not order_by:
        order_by = [{"order_by": "uuid", "order_direction": "desc"}]
    else:
        # Check if uuid is already in order_by
        has_uuid_order = any(
            order.get("order_by") == "uuid" for order in order_by
        )
        if not has_uuid_order:
            # Prepend UUID ordering for stable pagination
            order_by = [{"order_by": "uuid", "order_direction": "desc"}] + order_by

    # Build VQL query
    vql_query = VQLFilterBuilder.build_vql_query(
        where=where,
        select_columns=select_columns,
        order_by=order_by,
        limit=limit,
        company_config=company_config,
    )

    # Build job data matching ExportFileJobData structure
    job_data: Dict[str, Any] = {
        "s3_bucket": s3_bucket,
        "service": "company",
        "vql": vql_query,
        "workflow_uuid": workflow_id,
        "workflow_type": "company_export",
        "created_at": datetime.now().isoformat(),
    }

    # Add custom variables
    custom_vars_dict = {
        "source": "companies_db",
        "destination": "s3_csv",
    }
    if custom_vars:
        custom_vars_dict.update(custom_vars)
    job_data["custom_vars"] = custom_vars_dict

    # Build job node
    dag = [
        BaseDAGBuilder.build_job_node(
            uuid=workflow_id,
            job_title=f"Export Companies to CSV (limit: {limit or 'unlimited'})",
            job_type="export_csv_file",  # Connectra job type
            data=job_data,
            edges=[],
            retry_count=retry_count,
            retry_interval=retry_interval,
        )
    ]

    return dag
