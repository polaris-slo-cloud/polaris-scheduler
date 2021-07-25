import { MetricsSource, ObservableOrPromise, PolarisRuntime, ServiceLevelObjective, SloCompliance, SloMapping, SloOutput } from '@polaris-sloc/core';
import { ImageThroughputSloConfig } from '@rainbow-h2020/common-mappings';

/**
 * Implements the ImageThroughput SLO.
 *
 * ToDo: Change SloOutput type if necessary.
 */
export class ImageThroughputSlo implements ServiceLevelObjective<ImageThroughputSloConfig, SloCompliance> {
    sloMapping: SloMapping<ImageThroughputSloConfig, SloCompliance>;

    private metricsSource: MetricsSource;

    configure(
        sloMapping: SloMapping<ImageThroughputSloConfig, SloCompliance>,
        metricsSource: MetricsSource,
        polarisRuntime: PolarisRuntime,
    ): ObservableOrPromise<void> {
        this.sloMapping = sloMapping;
        this.metricsSource = metricsSource;

        // ToDo
    }

    evaluate(): ObservableOrPromise<SloOutput<SloCompliance>> {
        // ToDo
    }
}
