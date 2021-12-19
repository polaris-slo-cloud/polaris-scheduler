import { CustomStreamSightSloConfig } from '@rainbow-h2020/common-mappings';
import { MetricsSource, ObservableOrPromise, OrchestratorGateway, ServiceLevelObjective, SloCompliance, SloMapping, SloOutput } from '@polaris-sloc/core';

/**
 * Implements the CustomStreamSight SLO.
 *
 * ToDo: Change SloOutput type if necessary.
 */
export class CustomStreamSightSlo implements ServiceLevelObjective<CustomStreamSightSloConfig, SloCompliance> {
    sloMapping: SloMapping<CustomStreamSightSloConfig, SloCompliance>;

    private metricsSource: MetricsSource;

    configure(
        sloMapping: SloMapping<CustomStreamSightSloConfig, SloCompliance>,
        metricsSource: MetricsSource,
        orchestrator: OrchestratorGateway
    ): ObservableOrPromise<void> {
        this.sloMapping = sloMapping;
        this.metricsSource = metricsSource;

        // ToDo
    }

    evaluate(): ObservableOrPromise<SloOutput<SloCompliance>> {
        // ToDo
    }
}
