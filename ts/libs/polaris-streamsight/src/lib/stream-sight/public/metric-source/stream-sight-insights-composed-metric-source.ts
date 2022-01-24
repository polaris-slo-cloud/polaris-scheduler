import {
    ComposedMetricSourceBase,
    DataType,
    MetricsSource,
    OrchestratorGateway,
    Sample,
    TimeInstantQuery,
    TimeSeriesInstant,
    TimeSeriesSource,
} from '@polaris-sloc/core';
import { Observable, from, of as observableOf } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import { PolarisStreamSightConfig } from '../../../config';
import { RainbowStorageTimeSeriesSource } from '../../../rainbow-storage';
import { InsightTopologyManager } from '../../internal/insight-topology-manager';
import { StreamSightInsights, StreamSightInsightsMetric, StreamSightInsightsParams } from '../common';

/**
 * Specifies after how many value requests, the metric mapping should be checked again.
 *
 * See `StreamSightInsightsComposedMetricSource.maintainInsightTopology()`.
 */
const CHECK_INSIGHT_TOPOLOGY_INTERVAL = 100;

/**
 * A {@link ComposedMetricSource} that fetches insights from StreamSight.
 */
export class StreamSightInsightsComposedMetricSource extends ComposedMetricSourceBase<StreamSightInsights> {

    /** The query that fetches the insights from the RAINBOW storage. */
    protected query: TimeInstantQuery<Record<string, number>>;

    protected metricType = StreamSightInsightsMetric.instance;

    /** Total number times that this source has tried to query the composed metric. */
    private totalValueRequests = 0;

    /** Used to create/update the StreamSight insight topology for our metric. */
    private insightTopologyManager: InsightTopologyManager;

    constructor(
        protected metricParams: StreamSightInsightsParams,
        metricsSource: MetricsSource,
        orchestrator: OrchestratorGateway,
        protected streamSightConfig: PolarisStreamSightConfig,
    ) {
        super(metricsSource, orchestrator);
        this.insightTopologyManager = new InsightTopologyManager(streamSightConfig);
        const timeSeriesSource = metricsSource.getTimeSeriesSource(RainbowStorageTimeSeriesSource.fullName);
        this.query = this.createQuery(timeSeriesSource);
    }

    getValueStream(): Observable<Sample<StreamSightInsights>> {
        return this.getDefaultPollingInterval().pipe(
            switchMap(() => this.maintainInsightTopology()),
            switchMap(() => this.query.execute()),
            map(result => this.assembleComposedMetric(result.results)?.samples[0]),
        );
    }

    /**
     * Ensures the existence of the StreamSight insight topology for this composed metric on the first request and
     * subsequently on every nth request (e.g., on every 10th request).
     *
     * @returns An observable that emits (and subsequently completes) when the insight topology has been read/created/updated or if
     * this execution does not require a check.
     */
    private maintainInsightTopology(): Observable<void> {
        if (this.totalValueRequests++ % CHECK_INSIGHT_TOPOLOGY_INTERVAL === 0) {
            return from(this.insightTopologyManager.ensureInsightTopologyExists(this.metricParams)).pipe(
                map(() => undefined),
            );
        }
        return observableOf(undefined);
    }

    private createQuery(timeSeriesSource: TimeSeriesSource): TimeInstantQuery<Record<string, number>> {
        const metricName = this.insightTopologyManager.getInsightTopologyName(this.metricParams);
        return timeSeriesSource.select<Record<string, number>>('', metricName);
    }

    private assembleComposedMetric(results: TimeSeriesInstant<Record<string, number>>[]): TimeSeriesInstant<StreamSightInsights> {
        if (results?.length === 0) {
            return undefined;
        }
        const result = results[0];

        return {
            dataType: DataType.Object,
            metricName: this.metricType.metricTypeName,
            start: result.start,
            end: result.end,
            labels: {},
            samples: result.samples,
        };
    }

}
