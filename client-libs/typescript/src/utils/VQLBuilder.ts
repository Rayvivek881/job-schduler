/** VQL Filter Builder Utilities

Provides utilities for building VQL (Vivek Query Language) filter queries
used in export operations.
*/

import type {
  ElasticQuery,
  KeywordMatchOptions,
  RangeQueryOptions,
  TextMatchOptions,
  TextMatchQuery,
  VQLQuery,
  WhereStruct,
} from "../types/vql.types";

export class VQLFilterBuilder {
  /**
   * Build keyword_match filter
   */
  static keywordMatch(options: KeywordMatchOptions = {}): ElasticQuery {
    const keywordMatchFilter: ElasticQuery = {};
    if (options.must) {
      keywordMatchFilter.must = options.must;
    }
    if (options.mustNot) {
      keywordMatchFilter.must_not = options.mustNot;
    }
    return keywordMatchFilter;
  }

  /**
   * Build text_match filter
   */
  static textMatch(options: TextMatchOptions): {
    text_value: string;
    filter_key: string;
    search_type?: string;
    slop?: number;
    operator?: string;
    fuzzy?: boolean;
  } {
    const textMatchFilter: {
      text_value: string;
      filter_key: string;
      search_type?: string;
      slop?: number;
      operator?: string;
      fuzzy?: boolean;
    } = {
      text_value: options.textValue,
      filter_key: options.filterKey,
      search_type: options.searchType || "phrase",
    };

    if (options.slop !== undefined) {
      textMatchFilter.slop = options.slop;
    }
    if (options.operator) {
      textMatchFilter.operator = options.operator;
    }
    if (options.fuzzy) {
      textMatchFilter.fuzzy = options.fuzzy;
    }

    return textMatchFilter;
  }

  /**
   * Build range_query filter for numeric/date ranges
   */
  static rangeQuery(options: RangeQueryOptions = {}): ElasticQuery {
    const rangeFilter: ElasticQuery = {};
    if (options.must) {
      rangeFilter.must = options.must;
    }
    if (options.mustNot) {
      rangeFilter.must_not = options.mustNot;
    }
    return rangeFilter;
  }

  /**
   * Build complete VQL query structure
   */
  static buildVQLQuery(options: {
    where?: WhereStruct;
    selectColumns: string[];
    orderBy?: Array<{ order_by: string; order_direction: "asc" | "desc" }>;
    limit?: number;
    page?: number;
    cursor?: string[];
    companyConfig?: { populate?: boolean; select_columns?: string[] };
  }): VQLQuery {
    const vqlQuery: VQLQuery = {
      where: options.where || {
        text_matches: {},
        keyword_match: {},
        range_query: {},
      },
      select_columns: options.selectColumns,
    };

    if (options.orderBy) {
      vqlQuery.order_by = options.orderBy;
    }

    if (options.limit) {
      vqlQuery.limit = options.limit;
    }

    if (options.page) {
      vqlQuery.page = options.page;
    }

    if (options.cursor) {
      vqlQuery.cursor = options.cursor;
    }

    if (options.companyConfig) {
      vqlQuery.company_config = options.companyConfig;
    }

    return vqlQuery;
  }

  /**
   * Combine multiple filter types into where clause
   */
  static combineFilters(options: {
    keywordMatch?: ElasticQuery;
    textMatches?: Array<{
      text_value: string;
      filter_key: string;
      search_type?: string;
      slop?: number;
      operator?: string;
      fuzzy?: boolean;
    }>;
    rangeQuery?: ElasticQuery;
  }): WhereStruct {
    const where: WhereStruct = {};

    if (options.keywordMatch) {
      where.keyword_match = options.keywordMatch;
    }

    if (options.textMatches) {
      where.text_matches = {
        must: options.textMatches,
        must_not: [],
      };
    }

    if (options.rangeQuery) {
      where.range_query = options.rangeQuery;
    }

    return where;
  }
}
