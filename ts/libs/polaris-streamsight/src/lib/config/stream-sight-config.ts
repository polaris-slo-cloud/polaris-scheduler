
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
