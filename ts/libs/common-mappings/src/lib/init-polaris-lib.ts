import { PolarisRuntime } from '@polaris-sloc/core';
import { CustomStreamSightSloMapping } from './slo-mappings/custom-stream-sight.slo-mapping.prm';
import { ImageThroughputSloMapping } from './slo-mappings/image-throughput.slo-mapping';

/**
 * Initializes this library and registers its types with the transformer in the `PolarisRuntime`.
 */
export function initPolarisLib(polarisRuntime: PolarisRuntime): void {
    polarisRuntime.transformer.registerObjectKind(new ImageThroughputSloMapping().objectKind, ImageThroughputSloMapping);
    polarisRuntime.transformer.registerObjectKind(new CustomStreamSightSloMapping().objectKind, CustomStreamSightSloMapping);
}
