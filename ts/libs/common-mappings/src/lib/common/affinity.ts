
// Documentation partly copied from the Kubernetes project. See https://pkg.go.dev/k8s.io/api/core/v1

export interface K8sAffinityConfiguration {

    /** Describes node affinity scheduling rules for the pod. */
    nodeAffinity?: NodeAffinity;

    /** Describes pod affinity scheduling rules (e.g. co-locate this pod in the same node, zone, etc. as some other pod(s)). */
    podAffinity?: PodAffinity;

    /** Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)). */
    podAntiAffinity?: PodAffinity;

}

export interface NodeSelectorRequirement {

    /** The label key that the selector applies to. */
    key: string;

    /** Represents a key's relationship to a set of values. */
    operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist' | 'Gt' |  'Lt';

    /**
     * An array of string values. If the operator is In or NotIn
     * the values array must be non-empty. If the operator is Exists or DoesNotExist,
	 * the values array must be empty. If the operator is Gt or Lt, the values
	 * array must have a single element, which will be interpreted as an integer.
     */
    values?: string[];

}

/**
 * Specifies a single term for node selection.
 * A null or empty node selector term matches no objects. The requirements of them are ANDed.
 */
export interface NodeSelectorTerm {

    /** A list of node selector requirements by node's labels. */
    matchExpressions?: NodeSelectorRequirement[];

    /** A list of node selector requirements by node's fields. */
    matchFields?: NodeSelectorRequirement[];

}

/** Used for selecting a node. */
export interface NodeSelector {

    /** A list of node selector terms. The terms are ORed. */
    nodeSelectorTerms: NodeSelectorTerm[];
}

export interface PreferredSchedulingTerm {

    /**
     * Weight associated with matching the corresponding nodeSelectorTerm.
     *
     * @minimum 1
     * @maximum 100
     */
    weight: number;

    /** A node selector term, associated with the corresponding weight. */
    preference: NodeSelectorTerm;

}

export interface NodeAffinity {

    /** These conditions must be fulfilled by a node for successful scheduling. */
    requiredDuringSchedulingIgnoredDuringExecution?: NodeSelector;

    preferredDuringSchedulingIgnoredDuringExecution?: PreferredSchedulingTerm[];

}

export interface PodAffinity {

    requiredDuringSchedulingIgnoredDuringExecution?: Record<string, unknown>[];

    preferredDuringSchedulingIgnoredDuringExecution?: Record<string, unknown>[];

}
