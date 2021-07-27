import { NativeQueryBuilderBase, QueryError, TimeSeriesQuery, TimeSeriesQueryResultType } from '@polaris-sloc/core';
import { PolarisStreamSightConfig } from '../../config';
import { GetAnalyticsRequest } from '../../model';
import { RainbowStorageNativeQuery } from './rainbow-storage-native-query';

interface RainbowStorageQueryInfo {
    query: GetAnalyticsRequest;
    metricName: string;
}

/**
 * Builds queries for the RAINBOW Distributed Data Storage.
 *
 * Important: This can currently only handle select statements.
 */
export class RainbowStorageNativeQueryBuilder extends NativeQueryBuilderBase  {

    constructor(private config: PolarisStreamSightConfig) {
        super();
    }

    buildQuery(resultType: TimeSeriesQueryResultType): TimeSeriesQuery<any> {
        const query = this.buildRainbowStorageQuery();
        return new RainbowStorageNativeQuery(this.config, resultType, query.metricName, query.query);
    }

    private buildRainbowStorageQuery(): RainbowStorageQueryInfo {
        if (this.queryChainAfterSelect.length > 0) {
            throw new QueryError('RainbowStorageNativeQueryBuilder can currently only handle select statements.', this);
        }

        const metricName = `${this.selectSegment.appName}_${this.selectSegment.metricName}`;
        return {
            metricName,
            query: {
                key: [ metricName ],
            },
        };
    }

}
