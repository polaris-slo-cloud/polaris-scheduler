import { PolarisRuntime } from '@polaris-sloc/core';
import { ImageThroughputSloMappingSloMapping } from './slo-mappings/image-throughput-slo-mapping.slo-mapping';

/**
 * Initializes this library and registers its types with the transformer in the `PolarisRuntime`.
 */
export function initPolarisLib(polarisRuntime: PolarisRuntime): void {
    polarisRuntime.transformer.registerObjectKind(new ImageThroughputSloMappingSloMapping().objectKind, ImageThroughputSloMappingSloMapping);
}
