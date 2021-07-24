
/**
 * Used to request one or more monitoring metrics from the RAINBOW Distributed Data Storage.
 *
 * @see https://gitlab.com/rainbow-project1/rainbow-storage/-/tree/master#rest-api-examples
 */
export interface GetMetricsRequest {

    /**
     * The IDs of the metrics to get or the IDs of the entities for which to get all metrics.
     */
    metricId: string[];

    /**
     * Unix timestamp designating the starting point of the metrics.
     *
     * Optional, if `latest` is `true`.
     */
    from?: number;

    /**
     * Unix timestamp designating the end point of the metrics.
     *
     * Optional, if `latest` is `true`.
     */
    to?: number;

    /**
     * Only get the latest metric values.
     */
    latest: boolean;

}

/**
 * Used to request one or more analytics metrics from the RAINBOW Distributed Data Storage.
 *
 * @see https://gitlab.com/rainbow-project1/rainbow-storage/-/tree/master#rest-api-examples
 */
export interface GetAnalyticsRequest {

    /**
     * The keys, for which to retrieve the analytics data.
     */
    key: string[];

}

/**
 * The response to a `GetAnalyticsRequest`.
 */
export interface GetAnalyticsResponse {

    /** The analytics metrics that were found. */
    analytics: AnalyticsMetric[];

}

/**
 * Represents a single RAINBOW analytics metric value.
 */
export interface AnalyticsMetric {

    /** The key/name of the analytics metric. */
    key: string;

    /** The value of the metric. */
    val: number;

    /** The Unix timestamp when this metric was recorded. */
    timestamp: number;

}
