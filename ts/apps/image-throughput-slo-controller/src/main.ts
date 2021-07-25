import { KubeConfig } from '@kubernetes/client-node';
import { ImageThroughputSloMapping, ImageThroughputSloMappingSpec, initPolarisLib as initSloMappingsLib } from '@rainbow-h2020/common-mappings';
import { initPolarisKubernetes } from '@polaris-sloc/kubernetes';
import { initPrometheusQueryBackend } from '@polaris-sloc/prometheus';
import { interval } from 'rxjs';
import { ImageThroughputSlo } from './app/slo';
import { convertToNumber, getEnvironmentVariable } from './app/util/environment-var-helper';

// Load the KubeConfig and initialize the @polaris-sloc/kubernetes library.
const k8sConfig = new KubeConfig();
k8sConfig.loadFromDefault();
const polarisRuntime = initPolarisKubernetes(k8sConfig);

// Initialize the Prometheus query backend.
const promHost = getEnvironmentVariable('PROMETHEUS_HOST') || 'localhost';
const promPort = getEnvironmentVariable('PROMETHEUS_PORT', convertToNumber) || 9090;
initPrometheusQueryBackend(polarisRuntime, { host: promHost, port: promPort }, true);

// Initialize the used Polaris mapping libraries
initSloMappingsLib(polarisRuntime);

// ToDo: Initialize any additional libraries, e.g., composed metrics.

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
