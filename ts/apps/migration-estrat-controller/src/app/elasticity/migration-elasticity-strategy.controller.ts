import { ApiObjectMetadata, ElasticityStrategy, ElasticityStrategyExecutionError, Logger, ObjectKind, SloCompliance } from '@polaris-sloc/core';
import { PodTemplateContainer } from '@polaris-sloc/core';
import {
    DefaultStabilizationWindowTracker,
    OrchestratorClient,
    PodSpec,
    PolarisRuntime,
    SloComplianceElasticityStrategyControllerBase,
    SloTarget,
    StabilizationWindowTracker,
} from '@polaris-sloc/core';
import { MigrationElasticityStrategyConfig, MigrationElasticityStrategy, K8sAffinityConfiguration } from '@rainbow-h2020/common-mappings';

/** Tracked executions eviction interval of 20 minutes. */
const EVICTION_INTERVAL_MSEC = 20 * 60 * 1000;

class K8sPodSpec extends PodSpec {
    affinity?: K8sAffinityConfiguration;
}

/**
 * Controller for the MigrationElasticityStrategy.
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
        Logger.log('Executing elasticity strategy:', elasticityStrategy);
        const target = await this.loadTarget(elasticityStrategy);
        const podSpec = target.spec.template.spec as K8sPodSpec;
        let isOutsideStabilizationWindow: boolean;

        // At or below 100 we use the baseNodeAffinity,
        // above 100 we use the alternativeNodeAffinity.
        // We don't need to check the tolerance, because this has already been done by the superclass.
        if (elasticityStrategy.spec.sloOutputParams.currSloCompliancePercentage <= 100) {
            podSpec.affinity = elasticityStrategy.spec.staticConfig?.baseAffinity;
            isOutsideStabilizationWindow = this.stabilizationWindowTracker.isOutsideStabilizationWindowForScaleDown(elasticityStrategy)
        } else {
            podSpec.affinity = elasticityStrategy.spec.staticConfig?.alternativeAffinity;
            isOutsideStabilizationWindow = this.stabilizationWindowTracker.isOutsideStabilizationWindowForScaleUp(elasticityStrategy);
        }

        if (!isOutsideStabilizationWindow) {
            Logger.log(
                'Skipping scaling, because stabilization window has not yet passed for: ',
                elasticityStrategy,
            );
            return;
        }

        await this.orchClient.update(target);
        this.stabilizationWindowTracker.trackExecution(elasticityStrategy);
        Logger.log('Successfully scaled.', elasticityStrategy, JSON.stringify(podSpec.affinity, null, '  '));
    }

    onDestroy(): void {
        clearInterval(this.evictionInterval);
    }

    onElasticityStrategyDeleted(elasticityStrategy: MigrationElasticityStrategy): void {
        this.stabilizationWindowTracker.removeElasticityStrategy(elasticityStrategy);
    }

    private async loadTarget(elasticityStrategy: ElasticityStrategy<SloCompliance, SloTarget, MigrationElasticityStrategyConfig>): Promise<PodTemplateContainer> {
        const targetRef = elasticityStrategy.spec.targetRef;
        const queryApiObj = new PodTemplateContainer({
            objectKind: new ObjectKind({
                group: targetRef.group,
                version: targetRef.version,
                kind: targetRef.kind,
            }),
            metadata: new ApiObjectMetadata({
                namespace: elasticityStrategy.metadata.namespace,
                name: targetRef.name,
            }),
        });

        const ret = await this.orchClient.read(queryApiObj);
        if (!ret.spec?.template) {
            throw new ElasticityStrategyExecutionError('The SloTarget does not contain a pod template (spec.template field).', elasticityStrategy);
        }
        return ret;
    }
}
