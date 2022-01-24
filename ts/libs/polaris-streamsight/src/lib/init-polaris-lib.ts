import { PolarisRuntime } from '@polaris-sloc/core';
import { PolarisStreamSightConfig } from './config';
import { RainbowStorageTimeSeriesSource } from './rainbow-storage';
import { StreamSightInsightsComposedMetricSourceFactory } from './stream-sight';

/**
 * Initializes the StreamSightQueryBackend and registers it with the `PolarisRuntime`.
 *
 * @param runtime The `PolarisRuntime` instance.
 * @param config The configuration for accessing StreamSight and the RAINBOW Distributed Storage.
 * @param setAsDefaultSource If `true`, StreamSight will be set as the default `TimeSeriesSource`.
 */
 export function initStreamSightQueryBackend(runtime: PolarisRuntime, config: PolarisStreamSightConfig, setAsDefaultSource: boolean = false): void {
    console.log('Initializing StreamSightQueryBackend with config:', config);

    runtime.metricsSourcesManager.addTimeSeriesSource(new RainbowStorageTimeSeriesSource(config), setAsDefaultSource);

    const streamSightInsightsMetricFactory = new StreamSightInsightsComposedMetricSourceFactory(config);
    StreamSightInsightsComposedMetricSourceFactory.supportedSloTargetTypes.forEach(
        sloTargetType => runtime.metricsSourcesManager.addComposedMetricSourceFactory(streamSightInsightsMetricFactory, sloTargetType),
    );
}
