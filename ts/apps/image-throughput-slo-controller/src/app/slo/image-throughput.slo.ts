import { MetricsSource, ObservableOrPromise, PolarisRuntime, ServiceLevelObjective, SloCompliance, SloMapping, SloOutput } from '@polaris-sloc/core';
import { ImageThroughputSloConfig } from '@rainbow-h2020/common-mappings';
import { of as observableOf } from 'rxjs';

const THROUGHPUT_METRIC = 'overallframePerSecWindow';
const CPU_METRIC = 'overallCPU';

/**
 * Implements the ImageThroughput SLO.
 */
export class ImageThroughputSlo implements ServiceLevelObjective<ImageThroughputSloConfig, SloCompliance> {

    sloMapping: SloMapping<ImageThroughputSloConfig, SloCompliance>;

    private metricsSource: MetricsSource;
    private minCpuUsage: number;

    configure(
        sloMapping: SloMapping<ImageThroughputSloConfig, SloCompliance>,
        metricsSource: MetricsSource,
        polarisRuntime: PolarisRuntime,
    ): ObservableOrPromise<void> {
        this.sloMapping = sloMapping;
        this.metricsSource = metricsSource;
        this.minCpuUsage = sloMapping.spec.sloConfig.minCpuUsage ?? 70;
        return observableOf(null);
    }

    evaluate(): ObservableOrPromise<SloOutput<SloCompliance>> {
        return this.calculateSloCompliance().then(compliance => ({
            sloMapping: this.sloMapping,
            elasticityStrategyParams: {
                currSloCompliancePercentage: compliance,
            },
        }));
    }

    private async calculateSloCompliance(): Promise<number> {
        const throughputPerMinute = await this.getMetricValue('', THROUGHPUT_METRIC);
        let compliance = (this.sloMapping.spec.sloConfig.targetImagesPerMinute / throughputPerMinute) * 100;

        if (compliance > 100) {
            // If we need to increase resources, we will only do so, if the CPU usage is also high.
            const cpuUsage = await this.getMetricValue('', CPU_METRIC);
            if (cpuUsage < this.minCpuUsage) {
                compliance = 100;
            }
        }

        return compliance;
    }

    private async getMetricValue(appName: string, metricName: string): Promise<number> {
        const query = this.metricsSource.getTimeSeriesSource()
            .select('', metricName);
        const queryResult = await query.execute();

        if (queryResult.results.length === 0) {
            throw new Error(metricName + ' metric could not be read.');
        }
        return queryResult.results[0].samples[0].value;
    }
}
