"""DAG Template System

Provides template-based DAG creation for reusable patterns.
"""

import re
from typing import Any, Dict, List


class DAGTemplate:
    """Template system for DAG creation"""

    @staticmethod
    def process_template(
        template: Dict[str, Any], variables: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Process DAG template with variable substitution
        
        Replaces {variable} placeholders in template with actual values.
        Supports nested dictionaries and lists.
        
        Args:
            template: Template dictionary with {variable} placeholders
            variables: Dictionary of variable values
            
        Returns:
            Processed template with variables substituted
        """
        def replace_vars(obj: Any) -> Any:
            """Recursively replace variables in object"""
            if isinstance(obj, str):
                # Replace {variable} patterns
                for key, value in variables.items():
                    obj = obj.replace(f"{{{key}}}", str(value))
                return obj
            elif isinstance(obj, dict):
                return {k: replace_vars(v) for k, v in obj.items()}
            elif isinstance(obj, list):
                return [replace_vars(item) for item in obj]
            else:
                return obj

        return replace_vars(template)

    @staticmethod
    def create_etl_pipeline_template(
        date_str: str, bucket: str, config: Dict[str, Any]
    ) -> List[Dict[str, Any]]:
        """Create ETL pipeline DAG template
        
        Args:
            date_str: Date string for the pipeline
            bucket: S3 bucket name
            config: Configuration dictionary
            
        Returns:
            ETL pipeline DAG
        """
        base_uuid_prefix = f"etl-{date_str}"

        template = [
            {
                "uuid": f"{base_uuid_prefix}-extract",
                "job_title": f"Extract Data - {date_str}",
                "job_type": "etl",
                "data": {
                    "source_bucket": bucket,
                    "source_key": f"data/{date_str}/input.csv",
                    "format": config.get("format", "csv"),
                    "custom_vars": {
                        "date": date_str,
                        "region": config.get("region", "us-east-1"),
                    },
                },
                "retry_count": config.get("retry_count", 3),
                "retry_interval": config.get("retry_interval", 5),
                "edges": [f"{base_uuid_prefix}-transform"],
            },
            {
                "uuid": f"{base_uuid_prefix}-transform",
                "job_title": f"Transform Data - {date_str}",
                "job_type": "etl",
                "data": {
                    "transformations": config.get("transformations", []),
                    "output_format": config.get("output_format", "parquet"),
                },
                "retry_count": config.get("retry_count", 3),
                "retry_interval": config.get("retry_interval", 5),
                "edges": [f"{base_uuid_prefix}-load"],
            },
            {
                "uuid": f"{base_uuid_prefix}-load",
                "job_title": f"Load to Warehouse - {date_str}",
                "job_type": "etl",
                "data": {
                    "destination": config.get("warehouse", "warehouse"),
                    "table": config.get("table", "daily_data"),
                },
                "edges": [],
            },
        ]

        return template

    @staticmethod
    def get_template_variables(template: Dict[str, Any]) -> List[str]:
        """Extract variable names from template
        
        Finds all {variable} placeholders in template.
        
        Args:
            template: Template dictionary
            
        Returns:
            List of variable names found
        """
        variables = set()

        def extract_vars(obj: Any) -> None:
            """Recursively extract variable names"""
            if isinstance(obj, str):
                # Find all {variable} patterns
                matches = re.findall(r"\{([^}]+)\}", obj)
                variables.update(matches)
            elif isinstance(obj, dict):
                for value in obj.values():
                    extract_vars(value)
            elif isinstance(obj, list):
                for item in obj:
                    extract_vars(item)

        extract_vars(template)
        return sorted(list(variables))
