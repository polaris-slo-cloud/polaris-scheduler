import { NativeQueryBuilderFactoryFn, TimeSeriesSourceBase } from '@polaris-sloc/core';
import { PolarisStreamSightConfig } from '../../config';
import { RainbowStorageNativeQueryBuilder } from '../internal';

export class RainbowStorageTimeSeriesSource extends TimeSeriesSourceBase {

    readonly name = 'polaris-sloc.time-series-sources.RainbowStorage';

    constructor(protected config: PolarisStreamSightConfig) {
        super();
    }

    protected getNativeQueryBuilderFactory(): NativeQueryBuilderFactoryFn {
        return () => new RainbowStorageNativeQueryBuilder(this.config);
    }

}
