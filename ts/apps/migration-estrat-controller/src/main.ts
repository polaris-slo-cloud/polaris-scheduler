import { KubeConfig } from '@kubernetes/client-node';
import { MigrationElasticityStrategyKind, initPolarisLib as initMappingsLib } from '@rainbow-h2020/common-mappings';
import { Logger } from '@polaris-sloc/core';
import { initPolarisKubernetes } from '@polaris-sloc/kubernetes';
import { MigrationElasticityStrategyController } from './app/elasticity';

// Load the KubeConfig and initialize the @polaris-sloc/kubernetes library.
const k8sConfig = new KubeConfig();
k8sConfig.loadFromDefault();
const polarisRuntime = initPolarisKubernetes(k8sConfig);

// Initialize the used Polaris mapping libraries
initMappingsLib(polarisRuntime);

// Create an ElasticityStrategyManager and watch the supported elasticity strategy kinds.
const manager = polarisRuntime.createElasticityStrategyManager();
manager
    .startWatching({
        kindsToWatch: [{ kind: new MigrationElasticityStrategyKind(), controller: new MigrationElasticityStrategyController(polarisRuntime) }],
    })
    .catch(error => void Logger.error(error));
