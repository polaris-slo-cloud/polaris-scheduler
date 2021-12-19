
/**
 * Defines the desired target state of a StreamSight insight's value.
 */
export interface InsightTargetState {

    /** The insight, for which the state is defined. */
    insight: string;

    /**
     * The desired target value for the insight.
     *
     * By default we assume that a lower metric value is "better", e.g.,
     * for network latency a lower value is considered better than a higher value.
     * In this case, the following scaling approach is used:
     * - Above `targetValue + tolerance` we scale up/out.
     * - Below `targetValue - tolerance` we scale down/in
     *
     * This behavior can be inverted by setting the `higherIsBetter` property to `true`.
     *
     * @minimum 1
     */
    targetValue: number;

    /**
     * A tolerance around the target value.
     */
    tolerance: number;

    /**
     * (optional) If `true`, then a higher metric value is considered "better" and, thus,
     * the above/below rules of `targetValue` and `tolerance` are inverted.
     */
    higherIsBetter?: boolean;

}

/**
 * Defines a set of {@link InsightTargetState} objects, which are combined with an OR operator.
 */
export interface InsightDisjunction {

    /**
     * The states, which should be evaluated and combined with an OR operator.
     */
    disjuncts: InsightTargetState[]

}

/**
 * Defines an expression about StreamSight insights in CNF (Conjunctive Normal Form).
 *
 * The `conjuncts` property contains a list of disjunctions, which are evaluated and combined with an AND operator.
 *
 * Kubernetes CRDs do not seem to support recursive definitions (see https://github.com/kubernetes/kubernetes/issues/91669),
 * which means that we cannot support arbitrary nesting of AND and OR expression objects.
 * Thus, we have decided to require insight expression to be defined in a normalized form, i.e., CNF.
 */
export interface CNFInsightExpression {

    /**
     * The disjunction clauses, which are evaluated and combined with an AND operator.
     */
    conjuncts: InsightDisjunction[];

}
