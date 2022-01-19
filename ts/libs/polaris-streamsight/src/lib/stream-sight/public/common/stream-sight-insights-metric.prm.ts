import { ComposedMetricParams, ComposedMetricType, POLARIS_API } from '@polaris-sloc/core';

/**
 * Represents the value of the {@link StreamSightInsightsMetric} - a map of insight values.
 *
 * Each key is the name of an insight and its value is the latest value of the insight.
 */
export type StreamSightInsights = Record<string, number>;

/**
 * The parameters for retrieving the StreamSightInsights metric.
 */
export interface StreamSightInsightsParams extends ComposedMetricParams {

    /**
     * Defines the StreamSight streams that should be available for the insights.
     *
     * Each key in this object defines the name of the stream and its value is the definition of the stream.
     */
    streams: Record<string, string>;

    /**
     * Defines the insights that should be calculated.
     *
     * Each key in this object defines the name of an insight and its value specifies the query for it.
     */
    insights: Record<string, string>

}

/**
 * Represents the type of a generic cost efficiency metric.
 */
export class StreamSightInsightsMetric extends ComposedMetricType<StreamSightInsights, StreamSightInsightsParams> {
    /** The singleton instance of this type. */
    static readonly instance = new StreamSightInsightsMetric();

    readonly metricTypeName = POLARIS_API.METRICS_GROUP + '/v1/stream-sight';
}
