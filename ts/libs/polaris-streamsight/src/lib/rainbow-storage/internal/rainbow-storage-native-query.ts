import { DataType, Logger, Sample } from '@polaris-sloc/core';
import { PolarisQueryResult, QueryError, TimeSeries, TimeSeriesQuery, TimeSeriesQueryResultType } from '@polaris-sloc/core';
import { Observable, from as observableFrom } from 'rxjs';
import { IRestResponse, RestClient } from 'typed-rest-client/RestClient';
import { PolarisStreamSightConfig, getRainbowStorageBaseUrl } from '../../config';
import { AnalyticsMetric, GetAnalyticsRequest, GetAnalyticsResponse, RestRequestError } from '../../model';

const ANALYTICS_QUERY_PATH = '/analytics/get';

/**
 * A TimeSeriesQuery that contacts the RAINBOW Distributed Storage.
 */
export class RainbowStorageNativeQuery implements TimeSeriesQuery<any> {

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

    async execute(): Promise<PolarisQueryResult<TimeSeries<any>>> {
        const url = this.baseUrl + ANALYTICS_QUERY_PATH;
        const httpOptions: Record<string, string> = {
            // eslint-disable-next-line @typescript-eslint/naming-convention
            'Content-Type': 'application/json',
        };
        if (this.config.rainbowStorageAuthToken) {
            httpOptions['Authorization'] = this.config.rainbowStorageAuthToken;
        }

        let response: IRestResponse<GetAnalyticsResponse>;
        try {
            response = await this.client.create<GetAnalyticsResponse>(url, this.query, httpOptions);
        } catch (err) {
            const restError = new RestRequestError({ url, request: this.query, httpOptions, cause: err });
            throw new QueryError('Error executing RAINBOW Storage request.', this, restError);
        }

        if (response.statusCode !== 200 && response.statusCode !== 201) {
            const restError = new RestRequestError({ url, request: this.query, httpOptions, response });
            throw new QueryError('RAINBOW Storage returned an error.', this, restError);
        }

        const queryLog = {
            query: this.query,
            url,
            httpOptions,
            response,
        };
        Logger.log('RAINBOW Storage query successful:', JSON.stringify(queryLog, undefined, '  '));
        return this.transformQueryResponse(response.result);
    }

    toObservable(): Observable<PolarisQueryResult<any>> {
        return observableFrom(this.execute());
    }

    private transformQueryResponse(response: GetAnalyticsResponse): PolarisQueryResult<TimeSeries<number>> {
        const timeSeries = this.createTimeSeries();
        if (response?.analytics) {
            timeSeries.samples = response.analytics.map(sample => this.transformSample(sample));
        } else {
            timeSeries.samples = [];
        }
        return { results: [ timeSeries ] };
    }

    private createTimeSeries(): TimeSeries<number> {
        return {
            dataType: DataType.Float,
            metricName: this.metricName,
            labels: {},
            samples: null,
            start: null,
            end: null,
        };
    }

    private transformSample(sample: AnalyticsMetric): Sample<number> {
        return {
            timestamp: sample.timestamp,
            value: sample.val,
        };
    }

}
