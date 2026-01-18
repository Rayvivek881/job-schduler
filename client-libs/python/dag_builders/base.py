"""Base DAG Builder Class

Provides common functionality for all DAG builders including UUID generation,
workflow ID creation, and job node construction.
"""

import uuid
from datetime import datetime
from typing import Any, Dict, List, Optional


class BaseDAGBuilder:
    """Base class for all DAG builders with common utilities"""

    MAX_NODES_PER_REQUEST = 5000  # From constants.MaxNodesPerRequest

    @staticmethod
    def generate_uuid(prefix: Optional[str] = None) -> str:
        """Generate a unique UUID string
        
        Args:
            prefix: Optional prefix to prepend to UUID
            
        Returns:
            UUID string, optionally prefixed
        """
        uuid_str = uuid.uuid4().hex
        if prefix:
            return f"{prefix}-{uuid_str}"
        return uuid_str

    @staticmethod
    def generate_workflow_id(workflow_type: str, suffix: Optional[str] = None) -> str:
        """Generate a workflow ID with timestamp and optional suffix
        
        Args:
            workflow_type: Type of workflow (e.g., "contact-import", "email-export")
            suffix: Optional suffix (auto-generated UUID if None)
            
        Returns:
            Workflow ID string: "{workflow_type}-{timestamp}-{suffix}"
        """
        timestamp = datetime.now().strftime("%Y%m%d-%H%M%S")
        if suffix is None:
            suffix = uuid.uuid4().hex[:8]
        return f"{workflow_type}-{timestamp}-{suffix}"

    @staticmethod
    def build_job_node(
        uuid: str,
        job_title: str,
        job_type: str,
        data: Dict[str, Any],
        edges: Optional[List[str]] = None,
        retry_count: int = 3,
        retry_interval: int = 5,
        run_after: Optional[datetime] = None,
    ) -> Dict[str, Any]:
        """Build a job node dictionary for DAG submission
        
        Args:
            uuid: Unique job identifier
            job_title: Human-readable job title
            job_type: Type of job (e.g., "insert_csv_file", "export_csv_file")
            data: Job data payload (will be stored as JSONB)
            edges: List of target UUIDs this job depends on
            retry_count: Number of retry attempts
            retry_interval: Minutes between retries
            run_after: Optional datetime when job should run (defaults to now)
            
        Returns:
            Job node dictionary ready for API submission
        """
        job_node: Dict[str, Any] = {
            "uuid": uuid,
            "job_title": job_title,
            "job_type": job_type,
            "data": data,
            "retry_count": retry_count,
            "retry_interval": retry_interval,
            "edges": edges or [],
        }

        if run_after:
            job_node["run_after"] = run_after.isoformat()

        return job_node

    @staticmethod
    def validate_dag(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate DAG structure before submission
        
        Args:
            dag: List of job nodes
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        if not dag:
            return False, "DAG cannot be empty"

        if len(dag) > BaseDAGBuilder.MAX_NODES_PER_REQUEST:
            return (
                False,
                f"DAG exceeds maximum nodes limit ({BaseDAGBuilder.MAX_NODES_PER_REQUEST})",
            )

        # Check for duplicate UUIDs
        uuids = [node.get("uuid") for node in dag]
        if len(uuids) != len(set(uuids)):
            return False, "Duplicate UUIDs found in DAG"

        # Check all edge targets exist
        all_uuids = set(uuids)
        for node in dag:
            edges = node.get("edges", [])
            for edge_target in edges:
                if edge_target not in all_uuids:
                    return (
                        False,
                        f"Edge target '{edge_target}' not found in DAG nodes",
                    )

        # Validate required fields
        required_fields = ["uuid", "job_title", "job_type", "data"]
        for node in dag:
            for field in required_fields:
                if field not in node:
                    return False, f"Required field '{field}' missing in node {node.get('uuid', 'unknown')}"

        return True, None

    @staticmethod
    def add_custom_vars(
        data: Dict[str, Any], custom_vars: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Add custom variables to job data
        
        Args:
            data: Job data dictionary
            custom_vars: Custom variables to add
            
        Returns:
            Data dictionary with custom_vars merged in
        """
        if custom_vars:
            if "custom_vars" not in data:
                data["custom_vars"] = {}
            data["custom_vars"].update(custom_vars)
        return data
