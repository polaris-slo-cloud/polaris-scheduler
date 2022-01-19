import {
    DefaultStabilizationWindowTracker,
    OrchestratorClient,
    PolarisRuntime,
    SloComplianceElasticityStrategyControllerBase,
    SloTarget,
    StabilizationWindowTracker,
} from '@polaris-sloc/core';
import { MigrationElasticityStrategyConfig, MigrationElasticityStrategy } from '@rainbow-h2020/common-mappings';

/** Tracked executions eviction interval of 20 minutes. */
const EVICTION_INTERVAL_MSEC = 20 * 60 * 1000;

/**
 * Controller for the MigrationElasticityStrategy.
 *
 * ToDo:
 *  1. If you want to restrict the type of workloads that this elasticity strategy can be applied to,
 *     change the first generic parameter from `SloTarget` to the appropriate type.
 *  2. If your elasticity strategy input is not of type `SloCompliance`, change the definition of the controller class
 *     to extend `ElasticityStrategyController` instead of `SloComplianceElasticityStrategyControllerBase`.
 *  3. Implement the `execute()` method.
 *  4. Adapt `manifests/1-rbac.yaml` to include get and update permissions on all resources that you update in the orchestrator during `execute()`.
 */
export class MigrationElasticityStrategyController extends SloComplianceElasticityStrategyControllerBase<SloTarget, MigrationElasticityStrategyConfig> {
    /** The client for accessing orchestrator resources. */
    private orchClient: OrchestratorClient;

    /** Tracks the stabilization windows of the ElasticityStrategy instances. */
    private stabilizationWindowTracker: StabilizationWindowTracker<MigrationElasticityStrategy> = new DefaultStabilizationWindowTracker();

    private evictionInterval: NodeJS.Timeout;

    constructor(polarisRuntime: PolarisRuntime) {
        super();
        this.orchClient = polarisRuntime.createOrchestratorClient();

        this.evictionInterval = setInterval(() => this.stabilizationWindowTracker.evictExpiredExecutions(), EVICTION_INTERVAL_MSEC);
    }

    async execute(elasticityStrategy: MigrationElasticityStrategy): Promise<void> {
        // ToDo: Implement this method
    }

    onDestroy(): void {
        clearInterval(this.evictionInterval);
    }

    onElasticityStrategyDeleted(elasticityStrategy: MigrationElasticityStrategy): void {
        this.stabilizationWindowTracker.removeElasticityStrategy(elasticityStrategy);
    }
}
