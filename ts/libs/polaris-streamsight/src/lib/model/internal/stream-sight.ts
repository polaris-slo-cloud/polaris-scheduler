
/**
 * The StreamSight REST API path for insight topologies.
 * The name of the insight topology must be appended to this.
 */
export const STREAM_SIGHT_INSIGHTS_API_PATH = 'api/insights';

/**
 * Used to creat/edit an analytics insight topology.
 *
 * `POST api/insights/{{insight-topology-name}}`
 */
export interface CreateInsightTopologyRequest {

    /**
     * The stream and insight definition queries.
     *
     * @example
     * Queries: [
     *   'podstream: stream from storageLayer(periodicity=1000, metricID="cpu", entityType="POD", namespace="target-namespace", name="target-deployment-%" );',
     *   'avg_pod_cpu = COMPUTE AVG("cpu" FROM (podstream), 10s) EVERY 10s;'
     * ]
     */
    // eslint-disable-next-line @typescript-eslint/naming-convention
    Queries: string[];

}

/**
 * Response sent by StreamSight, when creating/editing an analytics insight.
 *
 * `POST api/insights/{{insight-topology-name}}`
 */
export interface CreateInsightTopologyResponse {

    /**
     * The status of the operation.
     */
    status: string;

}

/**
 * Response sent by StreamSight, when checking the status of an analytics insight topology.
 *
 * `GET api/insights/{{insight-topology-name}}`
 */
export interface GetInsightTopologyStatusResponse {

    /**
     * The status of this insight topology.
     */
    status: 'ACTIVE' | 'failure';

    /**
     * The queries that are part of this insight topology.
     */
    // eslint-disable-next-line @typescript-eslint/naming-convention
    Queries: string[];

    /**
     * Optional error message that explains the status.
     */
    message?: string;

}
