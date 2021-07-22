import { ObjectKind, PolarisType, SloCompliance, SloMappingBase, SloMappingInitData, SloMappingSpecBase, SloTarget, initSelf } from '@polaris-sloc/core';

/**
 * Represents the configuration options of the ImageThroughputSloMapping SLO.
 */
export interface ImageThroughputSloMappingSloConfig {

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
 * The spec type for the ImageThroughputSloMapping SLO.
 */
export class ImageThroughputSloMappingSloMappingSpec extends SloMappingSpecBase<
    // The SLO's configuration.
    ImageThroughputSloMappingSloConfig,
    // The output type of the SLO.
    SloCompliance,
    // The type of target(s) that the SLO can be applied to.
    SloTarget
> { }

/**
 * Represents an SLO mapping for the ImageThroughputSloMapping SLO, which can be used to apply and configure the ImageThroughputSloMapping SLO.
 */
export class ImageThroughputSloMappingSloMapping extends SloMappingBase<ImageThroughputSloMappingSloMappingSpec> {
    @PolarisType(() => ImageThroughputSloMappingSloMappingSpec)
    spec: ImageThroughputSloMappingSloMappingSpec;

    constructor(initData?: SloMappingInitData<ImageThroughputSloMappingSloMapping>) {
        super(initData);
        this.objectKind = new ObjectKind({
            group: 'slo.example.github.io', // ToDo: Replace the group with your own.
            version: 'v1',
            kind: 'ImageThroughputSloMappingSloMapping',
        });
        initSelf(this, initData);
    }
}
