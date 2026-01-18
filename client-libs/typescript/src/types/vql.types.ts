/** Type definitions for VQL (Vivek Query Language) structures */

/**
 * Filter order specification
 */
export interface FilterOrder {
  order_by: string;
  order_direction: "asc" | "desc";
}

/**
 * Text match structure for VQL queries
 */
export interface TextMatchStruct {
  text_value: string;
  filter_key: string;
  search_type?: string;
  slop?: number;
  operator?: string;
  fuzzy?: boolean;
}

/**
 * Text match query (must/must_not arrays)
 */
export interface TextMatchQuery {
  must?: TextMatchStruct[];
  must_not?: TextMatchStruct[];
}

/**
 * Elastic query structure (keyword match, range query)
 */
export interface ElasticQuery {
  must?: Record<string, unknown>;
  must_not?: Record<string, unknown>;
}

/**
 * Where clause structure for VQL
 */
export interface WhereStruct {
  text_matches?: TextMatchQuery;
  keyword_match?: ElasticQuery;
  range_query?: ElasticQuery;
}

/**
 * Company configuration for VQL queries
 */
export interface CompanyConfig {
  populate?: boolean;
  select_columns?: string[];
}

/**
 * Complete VQL query structure
 */
export interface VQLQuery {
  where?: WhereStruct;
  order_by?: FilterOrder[];
  cursor?: string[];
  select_columns: string[]; // Required for export jobs
  company_config?: CompanyConfig;
  page?: number;
  limit?: number;
}

/**
 * VQL filter builder options
 */
export interface KeywordMatchOptions {
  must?: Record<string, string[]>;
  mustNot?: Record<string, string[]>;
}

export interface TextMatchOptions {
  textValue: string;
  filterKey: string;
  searchType?: string;
  slop?: number;
  operator?: string;
  fuzzy?: boolean;
}

export interface RangeQueryOptions {
  must?: Record<string, unknown>;
  mustNot?: Record<string, unknown>;
}
