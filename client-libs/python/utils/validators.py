"""DAG Validation Utilities

Provides validation utilities for DAG structures before submission to the API.
"""

from typing import Any, Dict, List, Optional

from ..dag_builders.base import BaseDAGBuilder


class DAGValidator:
    """Validator for DAG structures"""

    @staticmethod
    def validate_dag(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate DAG structure
        
        This is a wrapper around BaseDAGBuilder.validate_dag for consistency.
        
        Args:
            dag: List of job nodes
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        return BaseDAGBuilder.validate_dag(dag)

    @staticmethod
    def validate_uuid_uniqueness(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate all UUIDs in DAG are unique
        
        Args:
            dag: List of job nodes
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        uuids = [node.get("uuid") for node in dag]
        unique_uuids = set(uuids)
        
        if len(uuids) != len(unique_uuids):
            duplicates = [uuid for uuid in uuids if uuids.count(uuid) > 1]
            return False, f"Duplicate UUIDs found: {set(duplicates)}"
        
        return True, None

    @staticmethod
    def validate_edges(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate all edge targets exist in DAG
        
        Args:
            dag: List of job nodes
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        all_uuids = {node.get("uuid") for node in dag}
        
        for node in dag:
            edges = node.get("edges", [])
            for edge_target in edges:
                if edge_target not in all_uuids:
                    return (
                        False,
                        f"Edge target '{edge_target}' from node '{node.get('uuid')}' not found in DAG",
                    )
        
        return True, None

    @staticmethod
    def validate_required_fields(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate all required fields are present
        
        Args:
            dag: List of job nodes
            
        Returns:
            Tuple of (is_valid, error_message)
        """
        required_fields = ["uuid", "job_title", "job_type", "data"]
        
        for node in dag:
            for field in required_fields:
                if field not in node:
                    return (
                        False,
                        f"Required field '{field}' missing in node {node.get('uuid', 'unknown')}",
                    )
        
        return True, None

    @staticmethod
    def validate_node_count(dag: List[Dict[str, Any]]) -> tuple[bool, Optional[str]]:
        """Validate node count is within limits
        
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
                f"DAG exceeds maximum nodes limit ({BaseDAGBuilder.MAX_NODES_PER_REQUEST}): {len(dag)} nodes",
            )
        
        return True, None
