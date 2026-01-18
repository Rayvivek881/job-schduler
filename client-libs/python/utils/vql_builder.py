"""VQL Filter Builder Utilities

Provides utilities for building VQL (Vivek Query Language) filter queries
used in export operations.
"""

from typing import Any, Dict, List, Optional


class VQLFilterBuilder:
    """Builder for constructing VQL query filters"""

    @staticmethod
    def keyword_match(
        must: Optional[Dict[str, List[str]]] = None,
        must_not: Optional[Dict[str, List[str]]] = None,
    ) -> Dict[str, Any]:
        """Build keyword_match filter
        
        Args:
            must: Fields that must match (field -> list of values)
            must_not: Fields that must not match
            
        Returns:
            keyword_match filter structure
        """
        keyword_match_filter: Dict[str, Any] = {}
        if must:
            keyword_match_filter["must"] = must
        if must_not:
            keyword_match_filter["must_not"] = must_not
        return keyword_match_filter

    @staticmethod
    def text_match(
        text_value: str,
        filter_key: str,
        search_type: str = "phrase",
        slop: Optional[int] = None,
        operator: Optional[str] = None,
        fuzzy: bool = False,
    ) -> Dict[str, Any]:
        """Build text_match filter
        
        Args:
            text_value: Text value to search for
            filter_key: Field to search in
            search_type: Type of search (phrase, match, etc.)
            slop: Word distance for phrase matching
            operator: Boolean operator (AND/OR)
            fuzzy: Enable fuzzy matching
            
        Returns:
            text_match filter structure
        """
        text_match_filter: Dict[str, Any] = {
            "text_value": text_value,
            "filter_key": filter_key,
            "search_type": search_type,
        }
        if slop is not None:
            text_match_filter["slop"] = slop
        if operator:
            text_match_filter["operator"] = operator
        if fuzzy:
            text_match_filter["fuzzy"] = fuzzy
        return text_match_filter

    @staticmethod
    def range_query(
        must: Optional[Dict[str, Any]] = None,
        must_not: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Build range_query filter for numeric/date ranges
        
        Args:
            must: Fields that must be in range
            must_not: Fields that must not be in range
            
        Returns:
            range_query filter structure
        """
        range_filter: Dict[str, Any] = {}
        if must:
            range_filter["must"] = must
        if must_not:
            range_filter["must_not"] = must_not
        return range_filter

    @staticmethod
    def build_vql_query(
        where: Optional[Dict[str, Any]] = None,
        select_columns: Optional[List[str]] = None,
        order_by: Optional[List[Dict[str, str]]] = None,
        limit: Optional[int] = None,
        page: Optional[int] = None,
        cursor: Optional[List[str]] = None,
        company_config: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Build complete VQL query structure
        
        Args:
            where: Where clause with filters
            select_columns: List of columns to select (required for exports)
            order_by: List of ordering rules [{"order_by": "field", "order_direction": "asc/desc"}]
            limit: Maximum number of records
            page: Page number for pagination
            cursor: Cursor for cursor-based pagination
            company_config: Company-specific configuration
            
        Returns:
            Complete VQL query dictionary
        """
        vql_query: Dict[str, Any] = {
            "where": where or {
                "text_matches": {},
                "keyword_match": {},
                "range_query": {},
            },
        }

        if select_columns:
            vql_query["select_columns"] = select_columns

        if order_by:
            vql_query["order_by"] = order_by

        if limit:
            vql_query["limit"] = limit

        if page:
            vql_query["page"] = page

        if cursor:
            vql_query["cursor"] = cursor

        if company_config:
            vql_query["company_config"] = company_config

        return vql_query

    @staticmethod
    def combine_filters(
        keyword_match: Optional[Dict[str, Any]] = None,
        text_matches: Optional[List[Dict[str, Any]]] = None,
        range_query: Optional[Dict[str, Any]] = None,
    ) -> Dict[str, Any]:
        """Combine multiple filter types into where clause
        
        Args:
            keyword_match: Keyword match filter
            text_matches: List of text match filters
            range_query: Range query filter
            
        Returns:
            Combined where clause
        """
        where: Dict[str, Any] = {}

        if keyword_match:
            where["keyword_match"] = keyword_match

        if text_matches:
            where["text_matches"] = {
                "must": text_matches,
                "must_not": [],
            }

        if range_query:
            where["range_query"] = range_query

        return where
