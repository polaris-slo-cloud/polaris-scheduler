import { DataType, Logger, Sample } from '@polaris-sloc/core';
import { PolarisQueryResult, QueryError, TimeSeries, TimeSeriesQuery, TimeSeriesQueryResultType } from '@polaris-sloc/core';
import { Observable, from as observableFrom } from 'rxjs';
import { IRestResponse, RestClient } from 'typed-rest-client/RestClient';
import { PolarisStreamSightConfig, getRainbowStorageBaseUrl } from '../../config';
import { AnalyticsMetric, GetAnalyticsRequest, GetAnalyticsResponse, RestRequestError } from '../../model';

const ANALYTICS_QUERY_PATH = '/analytics/get';

/**
 * A TimeSeriesQuery that reads analytics from the RAINBOW Distributed Storage.
 */
export class RainbowStorageNativeQuery implements TimeSeriesQuery<TimeSeries<Record<string, number>>> {

    private client: RestClient;
    private baseUrl: string;

    /**
     * Creates a new RainbowStorageNativeQuery.
     *
     * @param config The config used to connect to the RAINBOW Distributed Storage.
     * @param resultType The type of result that this query produces.
     * @param metricName The name of the metric that is returned by this query.
     * @param query The query object to be sent to the RAINBOW Distributed Storage.
     */
    constructor(
        private config: PolarisStreamSightConfig,
        public resultType: TimeSeriesQueryResultType,
        private metricName: string,
        private query: GetAnalyticsRequest,
    ) {
        this.client = new RestClient('polaris-query-backend');
        this.baseUrl = getRainbowStorageBaseUrl(config);
    }

    async execute(): Promise<PolarisQueryResult<TimeSeries<Record<string, number>>>> {
        const url = this.baseUrl + ANALYTICS_QUERY_PATH;
        // Send an empty object to get all currently stored analytics.
        const request = {};
        const httpOptions: Record<string, string> = {
            // eslint-disable-next-line @typescript-eslint/naming-convention
            'Content-Type': 'application/json',
        };
        if (this.config.rainbowStorageAuthToken) {
            httpOptions['Authorization'] = this.config.rainbowStorageAuthToken;
        }

        let response: IRestResponse<GetAnalyticsResponse>;
        try {
            response = await this.client.create<GetAnalyticsResponse>(url, request, httpOptions);
        } catch (err) {
            const restError = new RestRequestError({ url, request, httpOptions, cause: err });
            throw new QueryError('Error executing RAINBOW Storage request.', this, restError);
        }

        if (response.statusCode !== 200 && response.statusCode !== 201) {
            const restError = new RestRequestError({ url, request, httpOptions, response });
            throw new QueryError('RAINBOW Storage returned an error.', this, restError);
        }

        const queryLog = {
            request,
            url,
            httpOptions,
            response,
        };
        Logger.log('RAINBOW Storage query successful:', JSON.stringify(queryLog, undefined, '  '));
        return this.transformQueryResponse(response.result);
    }

    toObservable(): Observable<PolarisQueryResult<TimeSeries<Record<string, number>>>> {
        return observableFrom(this.execute());
    }

    private transformQueryResponse(response: GetAnalyticsResponse): PolarisQueryResult<TimeSeries<Record<string, number>>> {
        const timeSeries = this.createTimeSeries();
        if (response?.analytics?.length > 0) {
            const sample = this.transformAnalyticsList(response.analytics);
            timeSeries.samples = [ sample ];
            timeSeries.start = sample.timestamp;
            timeSeries.end = sample.timestamp;
        } else {
            timeSeries.samples = [];
        }
        return { results: [ timeSeries ] };
    }

    private createTimeSeries(): TimeSeries<Record<string, number>> {
        return {
            dataType: DataType.Object,
            metricName: this.metricName,
            labels: {},
            samples: null,
            start: null,
            end: null,
        };
    }

    private transformAnalyticsList(analytics: AnalyticsMetric[]): Sample<Record<string, number>> {
        const sample: Sample<Record<string, number>> = {
            timestamp: 0,
            value: {},
        };

        analytics.forEach(metric => {
            sample.value[metric.key] = metric.val;
            if (metric.timestamp > sample.timestamp) {
                sample.timestamp = metric.timestamp;
            }
        });

        return sample;
    }

}
