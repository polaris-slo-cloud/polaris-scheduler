import { IRestResponse, RestClient } from 'typed-rest-client';
import { PolarisStreamSightConfig, getStreamSightBaseUrl } from '../../config';
import { CreateInsightTopologyRequest, CreateInsightTopologyResponse, STREAM_SIGHT_INSIGHTS_API_PATH, StreamSightError } from '../../model';
import { StreamSightInsightsParams } from '../public/common';

const POD_NAMESPACE_PLACEHOLDER = '${namespace}';
const POD_NAME_PLACEHOLDER = '${podName}';

/**
 * Creates and updates StreamSight insight topologies.
 */
export class InsightTopologyManager {

    private client: RestClient;
    private baseUrl: string;

    constructor(private config: PolarisStreamSightConfig) {
        this.client = new RestClient('polaris-query-backend');
        this.baseUrl = getStreamSightBaseUrl(config);
    }

    /**
     * Ensures that the insight topology described by the `metricParams` exists and returns its name.
     */
    async ensureInsightTopologyExists(metricParams: StreamSightInsightsParams): Promise<string> {
        const insightName = this.getInsightTopologyName(metricParams);
        const url = `${this.baseUrl}/${STREAM_SIGHT_INSIGHTS_API_PATH}/${insightName}`;
        const req: CreateInsightTopologyRequest = {
            // eslint-disable-next-line @typescript-eslint/naming-convention
            Queries: this.assembleStreamSightQueries(metricParams),
        };

        let response: IRestResponse<CreateInsightTopologyResponse>;
        try {
            response = await this.client.create<CreateInsightTopologyResponse>(url, req);
        } catch (err) {
            throw new StreamSightError(undefined, req, err);
        }

        if (response.statusCode !== 200 && response.statusCode !== 201) {
            throw new StreamSightError(response, req);
        }
        if (response.result.status !== 'success') {
            throw new StreamSightError(response, req);
        }

        return insightName;
    }

    /**
     * Gets the name of the insight topology described by the `metricParams`.
     */
    getInsightTopologyName(metricParams: StreamSightInsightsParams): string {
        // Generate the name, based on the namespace and the SLO name.
        return `${metricParams.namespace}-${metricParams.owner.name}`
    }

    private assembleStreamSightQueries(metricParams: StreamSightInsightsParams): string[] {
        const streamKeys = Object.keys(metricParams.streams);
        const insightKeys = Object.keys(metricParams.insights);
        const queries: string[] = new Array(streamKeys.length + insightKeys.length);

        let i = 0;
        streamKeys.forEach(key => {
            const streamQuery = this.replacePlaceholders(metricParams.streams[key], metricParams);
            const query = `${key}: ${streamQuery}`;
            queries[i] = query;
            ++i;
        });
        insightKeys.forEach(key => {
            const insightQuery = this.replacePlaceholders(metricParams.insights[key], metricParams);
            const query = `${key} = ${insightQuery}`;
            queries[i] = query;
            ++i;
        });

        return queries;
    }

    private replacePlaceholders(query: string, metricParams: StreamSightInsightsParams): string {
        const namespace = metricParams.namespace;
        const podName = `${metricParams.sloTarget.name}-%`;
        let processedQuery = query.replace(POD_NAMESPACE_PLACEHOLDER, namespace);
        processedQuery = query.replace(POD_NAME_PLACEHOLDER, podName);
        return processedQuery;
    }

}
