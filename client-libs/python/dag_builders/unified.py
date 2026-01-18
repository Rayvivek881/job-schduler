"""Unified DAG Builder

Provides a unified interface for creating import/export DAGs for different services.
"""

from typing import Any, Dict, List, Optional

from .base import BaseDAGBuilder
from .company_export import create_company_export_dag
from .company_import import create_company_import_dag
from .contact_export import create_contact_export_dag
from .contact_import import create_contact_import_dag
from .email_export import create_email_export_dag


class ImportExportDAGBuilder:
    """Unified builder for import and export DAGs"""

    @staticmethod
    def create_import_dag(
        service: str,
        s3_bucket: str,
        s3_key: str,
        workflow_id: Optional[str] = None,
        retry_count: int = 3,
        retry_interval: int = 5,
        custom_vars: Optional[Dict[str, Any]] = None,
    ) -> List[Dict[str, Any]]:
        """Create import DAG for specified service
        
        Args:
            service: Service type ("contact" or "company")
            s3_bucket: S3 bucket containing CSV file
            s3_key: S3 key (path) of CSV file
            workflow_id: Optional workflow UUID
            retry_count: Number of retry attempts
            retry_interval: Minutes between retries
            custom_vars: Additional custom variables
            
        Returns:
            DAG definition ready for API submission
            
        Raises:
            ValueError: If service is not "contact" or "company"
        """
        if service == "contact":
            return create_contact_import_dag(
                s3_bucket=s3_bucket,
                s3_key=s3_key,
                workflow_id=workflow_id,
                retry_count=retry_count,
                retry_interval=retry_interval,
                custom_vars=custom_vars,
            )
        elif service == "company":
            return create_company_import_dag(
                s3_bucket=s3_bucket,
                s3_key=s3_key,
                workflow_id=workflow_id,
                retry_count=retry_count,
                retry_interval=retry_interval,
                custom_vars=custom_vars,
            )
        else:
            raise ValueError(
                f"Unsupported service: {service}. Must be 'contact' or 'company'"
            )

    @staticmethod
    def create_export_dag(
        service: str,
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
        """Create export DAG for specified service
        
        Args:
            service: Service type ("contact" or "company")
            s3_bucket: S3 bucket to store exported CSV
            select_columns: List of column names to export (required)
            workflow_id: Optional workflow UUID
            where: VQL where clause with filters
            order_by: List of ordering rules
            limit: Maximum number of records to export
            company_config: Company-specific configuration (only for company service)
            retry_count: Number of retry attempts
            retry_interval: Minutes between retries
            custom_vars: Additional custom variables
            
        Returns:
            DAG definition ready for API submission
            
        Raises:
            ValueError: If service is not "contact" or "company"
        """
        if service == "contact":
            return create_contact_export_dag(
                s3_bucket=s3_bucket,
                select_columns=select_columns,
                workflow_id=workflow_id,
                where=where,
                order_by=order_by,
                limit=limit,
                retry_count=retry_count,
                retry_interval=retry_interval,
                custom_vars=custom_vars,
            )
        elif service == "company":
            return create_company_export_dag(
                s3_bucket=s3_bucket,
                select_columns=select_columns,
                workflow_id=workflow_id,
                where=where,
                order_by=order_by,
                limit=limit,
                company_config=company_config,
                retry_count=retry_count,
                retry_interval=retry_interval,
                custom_vars=custom_vars,
            )
        else:
            raise ValueError(
                f"Unsupported service: {service}. Must be 'contact' or 'company'"
            )

    @staticmethod
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
        """Create email export enrichment DAG
        
        Convenience method that delegates to create_email_export_dag
        """
        return create_email_export_dag(
            s3_bucket=s3_bucket,
            s3_key=s3_key,
            workflow_id=workflow_id,
            output_s3_key=output_s3_key,
            email_api_url=email_api_url,
            email_api_key=email_api_key,
            batch_size=batch_size,
            retry_count=retry_count,
            retry_interval=retry_interval,
            custom_vars=custom_vars,
        )


# Factory function for convenience
def create_import_dag(
    service: str,
    s3_bucket: str,
    s3_key: str,
    **kwargs: Any,
) -> List[Dict[str, Any]]:
    """Factory function to create import DAG
    
    Convenience wrapper around ImportExportDAGBuilder.create_import_dag
    """
    return ImportExportDAGBuilder.create_import_dag(
        service=service, s3_bucket=s3_bucket, s3_key=s3_key, **kwargs
    )


def create_export_dag(
    service: str,
    s3_bucket: str,
    select_columns: List[str],
    **kwargs: Any,
) -> List[Dict[str, Any]]:
    """Factory function to create export DAG
    
    Convenience wrapper around ImportExportDAGBuilder.create_export_dag
    """
    return ImportExportDAGBuilder.create_export_dag(
        service=service,
        s3_bucket=s3_bucket,
        select_columns=select_columns,
        **kwargs,
    )
