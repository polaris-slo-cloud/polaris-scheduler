import { KubeConfig } from '@kubernetes/client-node';
import { Logger } from '@polaris-sloc/core';
import { initPolarisKubernetes } from '@polaris-sloc/kubernetes';
import { CustomStreamSightSloMapping, CustomStreamSightSloMappingSpec, initPolarisLib as initSloMappingsLib } from '@rainbow-h2020/common-mappings';
import { initStreamSightQueryBackend } from '@rainbow-h2020/polaris-streamsight';
import { interval } from 'rxjs';
import { CustomStreamSightSlo } from './app/slo';
import { convertToNumber, getEnvironmentVariable } from './app/util/environment-var-helper';

// Load the KubeConfig and initialize the @polaris-sloc/kubernetes library.
const k8sConfig = new KubeConfig();
k8sConfig.loadFromDefault();
// Really dirty hack to get around ERR_TLS_CERT_ALTNAME_INVALID, which occurs in hostNetwork mode.
k8sConfig.clusters = k8sConfig.clusters?.map(cluster => ({
    ...cluster,
    skipTLSVerify: true,
}));
const polarisRuntime = initPolarisKubernetes(k8sConfig);

// Initialize the RAINBOW StreamSight query backend.
const streamSightHost = getEnvironmentVariable('STREAM_SIGHT_HOST') || 'localhost';
const streamSightPort = getEnvironmentVariable('STREAM_SIGHT_PORT', convertToNumber);
const streamSightAuthToken = getEnvironmentVariable('STREAM_SIGHT_AUTH_TOKEN');
const rainbowStorageHost = getEnvironmentVariable('RAINBOW_STORAGE_HOST') || 'localhost';
const rainbowStoragePort = getEnvironmentVariable('RAINBOW_STORAGE_PORT', convertToNumber);
const rainbowStorageAuthToken = getEnvironmentVariable('RAINBOW_STORAGE_AUTH_TOKEN');
initStreamSightQueryBackend(
    polarisRuntime,
    {
        rainbowStorageHost,
        rainbowStoragePort,
        rainbowStorageAuthToken,
        streamSightHost,
        streamSightPort,
        streamSightAuthToken,
    },
    true,
);

// Initialize the used Polaris mapping libraries
initSloMappingsLib(polarisRuntime);

// Create an SloControlLoop and register the factories for the ServiceLevelObjectives it will handle
const sloControlLoop = polarisRuntime.createSloControlLoop();
sloControlLoop.microcontrollerFactory.registerFactoryFn(CustomStreamSightSloMappingSpec, () => new CustomStreamSightSlo());

// Create an SloEvaluator and start the control loop with an interval read from the SLO_CONTROL_LOOP_INTERVAL_MSEC environment variable (default is 20 seconds).
const sloEvaluator = polarisRuntime.createSloEvaluator();
const intervalMsec = getEnvironmentVariable('SLO_CONTROL_LOOP_INTERVAL_MSEC', convertToNumber) || 20000;
Logger.log(`Starting SLO control loop with an interval of ${intervalMsec} milliseconds.`);
sloControlLoop.start({
    evaluator: sloEvaluator,
    interval$: interval(intervalMsec),
});

// Create a WatchManager and watch the supported SLO mapping kinds.
const watchManager = polarisRuntime.createWatchManager();
watchManager.startWatchers([new CustomStreamSightSloMapping().objectKind], sloControlLoop.watchHandler).catch(error => void console.error(error));
