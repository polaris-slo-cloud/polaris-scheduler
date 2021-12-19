import {
    IndexByKey,
    ObjectKind,
    PolarisType,
    SloCompliance,
    SloMappingBase,
    SloMappingInitData,
    SloMappingSpecBase,
    SloTarget,
    initSelf,
} from '@polaris-sloc/core';
import { CNFInsightExpression } from '../common';

/**
 * Represents the configuration options of the CustomStreamSight SLO,
 * which allows specifying custom StreamSight queries.
 */
export interface CustomStreamSightSloConfig {

    /**
     * Defines the StreamSight streams that should be available for the insights.
     *
     * Each key in this object defines the name of the stream and its value is the definition of the stream.
     *
     * Within each stream definition, there are two placeholders that will be filled in by the SLO controller:
     * - `${namespace}`: The namespace, where the SloMapping is deployed.
     * - `${podName}`: A wildcard expression with the prefix of the pod names.
     */
    streams: IndexByKey<string>;

    /**
     * Defines the insights that can be used in the `conditions` below.
     *
     * Each key in this object defines the name of an insight and its value specifies the query for it.
     */
    insights: IndexByKey<string>

    /**
     * Defines the target state for the `insights`, i.e., the state in which the SLO should keep them,
     * in Conjunctive Normal Form (CNF).
     */
    targetState: CNFInsightExpression;

    /**
     * (optional) Specifies the tolerance around 100%, within which no scaling will be performed.
     *
     * For example, if tolerance is `10`, no scaling will be performed as long as the SloCompliance
     * is between `90` and `110`.
     *
     * @default 10
     */
    elasticityStrategyTolerance?: number;

}

/**
 * The spec type for the CustomStreamSight SLO.
 */
export class CustomStreamSightSloMappingSpec extends SloMappingSpecBase<
    // The SLO's configuration.
    CustomStreamSightSloConfig,
    // The output type of the SLO.
    SloCompliance,
    // The type of target(s) that the SLO can be applied to.
    SloTarget
> {}

/**
 * Represents the SLO mapping for the CustomStreamSight SLO, which allows specifying custom StreamSight queries in its configuration.
 */
export class CustomStreamSightSloMapping extends SloMappingBase<CustomStreamSightSloMappingSpec> {

    @PolarisType(() => CustomStreamSightSloMappingSpec)
    spec: CustomStreamSightSloMappingSpec;

    constructor(initData?: SloMappingInitData<CustomStreamSightSloMapping>) {
        super(initData);
        this.objectKind = new ObjectKind({
            group: 'slo.k8s.rainbow-h2020.eu',
            version: 'v1',
            kind: 'CustomStreamSightSloMapping',
        });
        initSelf(this, initData);
    }

}
