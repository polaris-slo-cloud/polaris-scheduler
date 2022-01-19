import { ElasticityStrategy, ElasticityStrategyKind, SloCompliance, SloTarget, initSelf } from '@polaris-sloc/core';
import { K8sAffinityConfiguration } from '../common';

/**
 * Configuration options for {@link MigrationElasticityStrategy}.
 *
 * This elasticity strategy allows moving the pods of a workload between two types of nodes, based on the SLO compliance.
 * This can be used, e.g., to normally run a workload on nodes of type A, but when a certain condition is true, move it to
 * nodes of type B.
 *
 * The node type selection is handled through affinities:
 * - `baseAffinity` is applied when the SLO Compliance is below `100 - tolerance`.
 * - `alternativeAffinity` is applied when the SLO Compliance is above `100 + tolerance`.
 *
 * If the SLO Compliance is between `100 - tolerance` and `100 + tolerance`, no change to the current situation is made.
 * The `tolerance` refers to the `tolerance` property of the `SloCompliance` object.
 */
export interface MigrationElasticityStrategyConfig {

    /**
     * This is applied when the SLO Compliance is below `100 - tolerance`.
     *
     * Not setting this, means that any affinity will be cleared when switching to the `baseAffinity`.
     */
    baseAffinity?: K8sAffinityConfiguration;

    /**
     * This is applied when the SLO Compliance is above `100 + tolerance`.
     *
     * Not setting this, means that any affinity will be cleared when switching to the `alternativeAffinity`.
     */
    alternativeAffinity?: K8sAffinityConfiguration;

}

/**
 * Denotes the elasticity strategy kind for the {@link MigrationElasticityStrategy}.
 *
 * See {@link MigrationElasticityStrategyConfig} for details on this elasticity strategy.
 */
export class MigrationElasticityStrategyKind extends ElasticityStrategyKind<SloCompliance, SloTarget> {
    constructor() {
        super({
            group: 'elasticity.k8s.rainbow-h2020.eu',
            version: 'v1',
            kind: 'MigrationElasticityStrategy',
        });
    }
}

/**
 * Defines the `MigrationElasticityStrategy`.
 *
 * See {@link MigrationElasticityStrategyConfig} for details on this elasticity strategy.
 */
export class MigrationElasticityStrategy extends ElasticityStrategy<SloCompliance, SloTarget, MigrationElasticityStrategyConfig> {
    constructor(initData?: Partial<MigrationElasticityStrategy>) {
        super(initData);
        this.objectKind = new MigrationElasticityStrategyKind();
        initSelf(this, initData);
    }
}
