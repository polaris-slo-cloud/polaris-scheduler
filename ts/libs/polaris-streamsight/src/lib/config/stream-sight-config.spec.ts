/* eslint-disable prefer-arrow/prefer-arrow-functions */
import { PolarisStreamSightConfig, getRainbowStorageBaseUrl } from './stream-sight-config';

describe('stream-sight-config', () => {

    describe('getRainbowStorageBaseUrl()', () => {

        function getStreamSightConfig(): PolarisStreamSightConfig {
            return {
                rainbowStorageHost: 'test',
            };
        }

        it('should return an http URL if TLS is not set', () => {
            const config = getStreamSightConfig();
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('http://test:50000');
        });

        it('should return an https URL if TLS is true', () => {
            const config = getStreamSightConfig();
            config.useTLS = true;
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('https://test:50000');
        });

        it('should return a URL with the configured port', () => {
            const config = getStreamSightConfig();
            config.rainbowStoragePort = 4711;
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('http://test:4711');
        });

    });

});
