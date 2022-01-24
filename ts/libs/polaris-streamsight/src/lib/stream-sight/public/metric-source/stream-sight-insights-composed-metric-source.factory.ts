import {
    ComposedMetricSource,
    ComposedMetricSourceFactory,
    MetricsSource,
    ObjectKind,
    OrchestratorGateway,
} from '@polaris-sloc/core';
import { PolarisStreamSightConfig } from '../../../config';
import { StreamSightInsights, StreamSightInsightsMetric, StreamSightInsightsParams } from '../common';
import { StreamSightInsightsComposedMetricSource } from './stream-sight-insights-composed-metric-source';

/**
 * Generic factory for creating {@link StreamSightComposedMetricSource} instances.
 */
export class StreamSightInsightsComposedMetricSourceFactory
    implements ComposedMetricSourceFactory<StreamSightInsightsMetric, StreamSightInsights, StreamSightInsightsParams>  {

    /**
     * The list of supported `SloTarget` types.
     *
     * This list can be used for registering an instance of this factory for each supported
     * `SloTarget` type with the `MetricsSourcesManager`. This registration must be done if the metric source should execute in the current process,
     * i.e., metric source instances can be requested through `MetricSource.getComposedMetricSource()`.
     *
     * When creating a composed metric controller, the list of compatible `SloTarget` types is determined by
     * the `ComposedMetricMapping` type.
     */
     static supportedSloTargetTypes: ObjectKind[] = [
        new ObjectKind({
            group: 'apps',
            version: 'v1',
            kind: 'Deployment',
        }),
        new ObjectKind({
            group: 'apps',
            version: 'v1',
            kind: 'StatefulSet',
        }),
        new ObjectKind({
            group: 'apps',
            version: 'v1',
            kind: 'ReplicaSet',
        }),
        new ObjectKind({
            group: 'apps',
            version: 'v1',
            kind: 'DaemonSet',
        }),
    ];

    readonly metricType = StreamSightInsightsMetric.instance;

    readonly metricSourceName = `${StreamSightInsightsMetric.instance.metricTypeName}/stream-sight-insights`;

    constructor(protected streamSightConfig: PolarisStreamSightConfig) {}

    createSource(
        params: StreamSightInsightsParams,
        metricsSource: MetricsSource,
        orchestrator: OrchestratorGateway,
    ): ComposedMetricSource<StreamSightInsights> {
        return new StreamSightInsightsComposedMetricSource(params, metricsSource, orchestrator, this.streamSightConfig);
    }

}
