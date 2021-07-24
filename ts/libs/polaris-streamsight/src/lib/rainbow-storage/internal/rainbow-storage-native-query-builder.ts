import { NativeQueryBuilderBase, TimeSeriesQuery, TimeSeriesQueryResultType } from '@polaris-sloc/core';
import { PolarisStreamSightConfig } from '../../config';

export class RainbowStorageNativeQueryBuilder extends NativeQueryBuilderBase  {

    constructor(private config: PolarisStreamSightConfig) {
        super();
    }

    buildQuery(resultType: TimeSeriesQueryResultType): TimeSeriesQuery<any> {
        throw new Error('Method not implemented.');
    }

}
