
/**
 * Used to request one or more monitoring metrics from the RAINBOW Distributed Data Storage.
 *
 * @see https://gitlab.com/rainbow-project1/rainbow-storage/-/tree/master#rest-api-examples
 */
export interface GetMetricsRequest {

    /**
     * The IDs of the metrics to get.
     */
    metricID: string[];

    /**
     * The IDs of the entities for which to get the metrics.
     *
     * If this is not specified, the metrics are retrieved for all entities.
     */
    entityID?: string[];

    /**
     * The pod names for which to get the metrics. Use `%` as a wildcard.
     */
    podName?: string[];

    /**
     * The namespace of the pods, for which to get the metrics.
     */
    podNamespace?: string[];

    /**
     * Allows limiting the results to specific containers.
     */
    containerName?: string[];

    /**
     * Limit the results to metrics from specific nodes.
     *
     * An empty array is equivalent to all nodes.
     */
    nodes?: string[];

    /**
     * Unix timestamp designating the starting point of the metrics.
     *
     * If not set, the latest metrics are retrieved.
     */
    from?: number;

    /**
     * Unix timestamp designating the end point of the metrics.
     *
     * If not set, the latest metrics are retrieved.
     */
    to?: number;

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

    /**
     * (optional) Limits the results to metrics from specific nodes.
     *
     * An empty array is equivalent to all nodes.
     */
    nodes?: string[];

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
