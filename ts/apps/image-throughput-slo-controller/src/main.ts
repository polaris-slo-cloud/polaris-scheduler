import { KubeConfig } from '@kubernetes/client-node';
import { initPolarisKubernetes } from '@polaris-sloc/kubernetes';
import { ImageThroughputSloMapping, ImageThroughputSloMappingSpec, initPolarisLib as initSloMappingsLib } from '@rainbow-h2020/common-mappings';
import { initStreamSightQueryBackend } from '@rainbow-h2020/polaris-streamsight';
import { interval } from 'rxjs';
import { ImageThroughputSlo } from './app/slo';
import { convertToNumber, getEnvironmentVariable } from './app/util/environment-var-helper';

// Load the KubeConfig and initialize the @polaris-sloc/kubernetes library.
const k8sConfig = new KubeConfig();
k8sConfig.loadFromDefault();
const polarisRuntime = initPolarisKubernetes(k8sConfig);

// Initialize the RAINBOW StreamSight query backend.
const streamSightHost = getEnvironmentVariable('STREAM_SIGHT_HOST') || 'localhost';
const streamSightPort = getEnvironmentVariable('STREAM_SIGHT_PORT', convertToNumber);
const rainbowStorageHost = getEnvironmentVariable('RAINBOW_STORAGE_HOST') || 'localhost';
const rainbowStoragePort = getEnvironmentVariable('RAINBOW_STORAGE_PORT', convertToNumber);
initStreamSightQueryBackend(polarisRuntime, { rainbowStorageHost, rainbowStoragePort, streamSightHost, streamSightPort }, true);

// Initialize the used Polaris mapping libraries
initSloMappingsLib(polarisRuntime);

// Create an SloControlLoop and register the factories for the ServiceLevelObjectives it will handle
const sloControlLoop = polarisRuntime.createSloControlLoop();
sloControlLoop.microcontrollerFactory.registerFactoryFn(ImageThroughputSloMappingSpec, () => new ImageThroughputSlo());

// Create an SloEvaluator and start the control loop with an interval of 20 seconds.
const sloEvaluator = polarisRuntime.createSloEvaluator();
sloControlLoop.start({
    evaluator: sloEvaluator,
    interval$: interval(20000),
});

// Create a WatchManager and watch the supported SLO mapping kinds.
const watchManager = polarisRuntime.createWatchManager();
watchManager.startWatchers([new ImageThroughputSloMapping().objectKind], sloControlLoop.watchHandler).catch(error => void console.error(error));
