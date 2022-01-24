
const STREAM_SIGHT_DEFAULT_PORT = 5000;
const RAINBOW_STORAGE_DEFAULT_PORT = 50000;

/**
 * Configuration for the Polaris StreamSight query backend.
 */
export interface PolarisStreamSightConfig {

    /**
     * Determines whether to use an HTTPS connection.
     *
     * Default: `false`
     */
    useTLS?: boolean;

    /**
     * The host, where the RAINBOW Distributed Storage can be reached.
     */
    rainbowStorageHost: string;

    /**
     * The port, where the RAINBOW Distributed Storage is listening
     *
     * Default: 50000
     */
    rainbowStoragePort?: number;

    /**
     * The token used to authenticate to the RAINBOW Distributed Storage.
     */
    rainbowStorageAuthToken?: string;

    /**
     * The host, where the StreamSight analytics service can be reached.
     */
    streamSightHost: string;

    /**
     * The port, where the StreamSight analytics service is listening.
     */
    streamSightPort?: number;

    /**
     * The token used to authenticate to StreamSight.
     */
    streamSightAuthToken?: string;

    /**
     * Number of milliseconds before a request goes into timeout.
     */
    timeout?: number;

}

/**
 * @returns The RAINBOW Distributed Storage base URL, based on the specified `config`.
 */
export function getRainbowStorageBaseUrl(config: PolarisStreamSightConfig): string {
    const protocol = config.useTLS ? 'https' : 'http';
    const port = config.rainbowStoragePort || RAINBOW_STORAGE_DEFAULT_PORT;
    return `${protocol}://${config.rainbowStorageHost}:${port}`;
}

/**
 * @returns The StreamSight base URL, based on the specified `config`.
 */
export function getStreamSightBaseUrl(config: PolarisStreamSightConfig): string {
    const protocol = config.useTLS ? 'https' : 'http';
    const port = config.streamSightPort || STREAM_SIGHT_DEFAULT_PORT;
    return `${protocol}://${config.streamSightHost}:${port}`;
}
