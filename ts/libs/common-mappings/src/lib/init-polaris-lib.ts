import { PolarisRuntime } from '@polaris-sloc/core';
import { ImageThroughputSloMapping } from './slo-mappings/image-throughput.slo-mapping';

/**
 * Initializes this library and registers its types with the transformer in the `PolarisRuntime`.
 */
export function initPolarisLib(polarisRuntime: PolarisRuntime): void {
    polarisRuntime.transformer.registerObjectKind(new ImageThroughputSloMapping().objectKind, ImageThroughputSloMapping);
}
