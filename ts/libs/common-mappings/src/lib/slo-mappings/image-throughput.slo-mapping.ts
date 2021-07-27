import { ObjectKind, PolarisType, SloCompliance, SloMappingBase, SloMappingInitData, SloMappingSpecBase, SloTarget, initSelf } from '@polaris-sloc/core';

/**
 * Represents the configuration options of the ImageThroughput SLO.
 */
export interface ImageThroughputSloConfig {

    /**
     * The desired number of images that should be processed per minute.
     */
    targetImagesPerMinute: number;

    /**
     * The minimum CPU usage percentage that must be achieved before scaling out on a
     * too low targetImagesPerMinute rate.
     *
     * Default: 70
     */
    minCpuUsage?: number;

}

/**
 * The spec type for the ImageThroughput SLO.
 */
export class ImageThroughputSloMappingSpec extends SloMappingSpecBase<
    // The SLO's configuration.
    ImageThroughputSloConfig,
    // The output type of the SLO.
    SloCompliance,
    // The type of target(s) that the SLO can be applied to.
    SloTarget
> { }

/**
 * Represents an SLO mapping for the ImageThroughput SLO, which can be used to apply and configure the ImageThroughput SLO.
 */
export class ImageThroughputSloMapping extends SloMappingBase<ImageThroughputSloMappingSpec> {
    @PolarisType(() => ImageThroughputSloMappingSpec)
    spec: ImageThroughputSloMappingSpec;

    constructor(initData?: SloMappingInitData<ImageThroughputSloMapping>) {
        super(initData);
        this.objectKind = new ObjectKind({
            group: 'slo.k8s.rainbow-h2020.eu',
            version: 'v1',
            kind: 'ImageThroughputSloMapping',
        });
        initSelf(this, initData);
    }
}
