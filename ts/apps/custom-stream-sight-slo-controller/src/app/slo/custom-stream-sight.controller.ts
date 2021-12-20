import {
    ComposedMetricSource,
    MetricsSource,
    ObservableOrPromise,
    OrchestratorGateway,
    ServiceLevelObjective,
    SloCompliance,
    SloMapping,
    SloOutput,
    createOwnerReference,
} from '@polaris-sloc/core';
import { CNFInsightExpression, CustomStreamSightSloConfig, InsightDisjunction, InsightTargetState } from '@rainbow-h2020/common-mappings';
import { StreamSightInsights, StreamSightInsightsMetric, StreamSightInsightsParams } from '@rainbow-h2020/polaris-streamsight';
import { of as observableOf } from 'rxjs';

const DEFAULT_ELASTICITY_STRATEGY_TOLERANCE = 10;

export class StreamSightSloError extends Error {

    constructor(
        message: string,
        public sloMapping: SloMapping<CustomStreamSightSloConfig, SloCompliance>,
        public insights?: StreamSightInsights,
    ) {
        super(message);
    }

}

/**
 * Implements the CustomStreamSight SLO.
 */
export class CustomStreamSightSlo implements ServiceLevelObjective<CustomStreamSightSloConfig, SloCompliance> {

    sloMapping: SloMapping<CustomStreamSightSloConfig, SloCompliance>;

    /** The SloMappings elasticityStrategyTolerance divided by 100. */
    private elasticityStrategyToleranceFloat: number;

    private streamSightMetricSource: ComposedMetricSource<StreamSightInsights>;

    configure(
        sloMapping: SloMapping<CustomStreamSightSloConfig, SloCompliance>,
        metricsSource: MetricsSource,
        orchestrator: OrchestratorGateway,
    ): ObservableOrPromise<void> {
        this.sloMapping = sloMapping;
        this.elasticityStrategyToleranceFloat = (sloMapping.spec.sloConfig.elasticityStrategyTolerance || DEFAULT_ELASTICITY_STRATEGY_TOLERANCE) / 100;

        const streamSightParams: StreamSightInsightsParams = {
            namespace: sloMapping.metadata.namespace,
            sloTarget: sloMapping.spec.targetRef,
            owner: createOwnerReference(sloMapping),
            streams: sloMapping.spec.sloConfig.streams,
            insights: sloMapping.spec.sloConfig.insights,
        };
        this.streamSightMetricSource = metricsSource.getComposedMetricSource(StreamSightInsightsMetric.instance, streamSightParams);

        return observableOf(undefined);
    }

    evaluate(): ObservableOrPromise<SloOutput<SloCompliance>> {
        return this.calculateSloCompliance()
            .then(sloCompliance => ({
                sloMapping: this.sloMapping,
                elasticityStrategyParams: {
                    currSloCompliancePercentage: sloCompliance,
                    tolerance: this.sloMapping.spec.sloConfig.elasticityStrategyTolerance,
                },
            }));
    }

    private async calculateSloCompliance(): Promise<number> {
        const sample = await this.streamSightMetricSource.getCurrentValue().toPromise();
        if (!sample) {
            throw new StreamSightSloError('Could not retrieve insight values from RAINBOW storage.', this.sloMapping);
        }
        const insights = sample.value;

        const complianceFloat = this.evaluateCNF(this.sloMapping.spec.sloConfig.targetState, insights);
        return complianceFloat * 100;
    }

    /**
     * Evaluates the CNF expression and returns the product of its conjuncts.
     */
    private evaluateCNF(cnf: CNFInsightExpression, insights: StreamSightInsights): number {
        let product = 1;
        cnf.conjuncts.forEach(disjunction => {
            const compliance = this.evaluateDisjunction(disjunction, insights);
            if (compliance > 0) {
                product *= compliance;
            }
        });
        return product;
    }

    /**
     * Evaluates all `InsightTargetStates` in the `disjunction` and determines the highest and the lowest SLO compliance values.
     *
     * @returns The highest SLO compliance value, if it exceeds the `elasticityStrategyTolerance`, otherwise
     * the lowest SLO compliance value, if it exceeds the `elasticityStrategyTolerance`, or
     * `1.0`, if both values are within the tolerance.
     */
    private evaluateDisjunction(disjunction: InsightDisjunction, insights: StreamSightInsights): number {
        let lowestCompliance = 1;
        let highestCompliance = 1;

        disjunction.disjuncts.forEach(targetState => {
            const compliance = this.evaluateInsight(targetState, insights);
            if (compliance < lowestCompliance) {
                lowestCompliance = compliance;
            }
            if (compliance > highestCompliance) {
                highestCompliance = compliance;
            }
        });

        const upperBound = 1 + this.elasticityStrategyToleranceFloat;
        const lowerBound = 1 - this.elasticityStrategyToleranceFloat;
        if (highestCompliance > upperBound) {
            return highestCompliance;
        }
        if (lowestCompliance < lowerBound) {
            return lowestCompliance;
        }
        return 1;
    }

    /**
     * Evaluates the insight described by `targetState` and returns a floating point SLO compliance percentage (i.e., 100% = `1.0`).
     *
     * If the insight's value is within the targetState's tolerance, the return value is 1.0 (100% compliance).
     */
    private evaluateInsight(targetState: InsightTargetState, insights: StreamSightInsights): number {
        const insightValue = insights[targetState.insight];
        if (insightValue === undefined) {
            throw new StreamSightSloError(`Insight ${targetState.insight} not found.`, this.sloMapping, insights);
        }

        const upperBound = targetState.targetValue + targetState.tolerance;
        const lowerBound = targetState.targetValue - targetState.tolerance;
        if (insightValue <= upperBound && insightValue >= lowerBound) {
            return 1;
        }

        let sloCompliance: number;
        if (targetState.higherIsBetter) {
            sloCompliance = targetState.targetValue / (insightValue || 1);
        } else {
            sloCompliance = insightValue / targetState.targetValue;
        }
        return Math.abs(sloCompliance);
    }
}
