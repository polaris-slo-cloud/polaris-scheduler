/* eslint-disable prefer-arrow/prefer-arrow-functions */
import { PolarisStreamSightConfig, getRainbowStorageBaseUrl, getStreamSightBaseUrl } from './stream-sight-config';

describe('stream-sight-config', () => {

    function getStreamSightConfig(): PolarisStreamSightConfig {
        return {
            rainbowStorageHost: 'storage.test',
            streamSightHost: 'stream-sight.test',
        };
    }

    describe('getRainbowStorageBaseUrl()', () => {

        it('should return an http URL if TLS is not set', () => {
            const config = getStreamSightConfig();
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('http://storage.test:50000');
        });

        it('should return an https URL if TLS is true', () => {
            const config = getStreamSightConfig();
            config.useTLS = true;
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('https://storage.test:50000');
        });

        it('should return a URL with the configured port', () => {
            const config = getStreamSightConfig();
            config.rainbowStoragePort = 4711;
            const url = getRainbowStorageBaseUrl(config);
            expect(url).toEqual('http://storage.test:4711');
        });

    });

    describe('getStreamSightBaseUrl()', () => {

        it('should return an http URL if TLS is not set', () => {
            const config = getStreamSightConfig();
            const url = getStreamSightBaseUrl(config);
            expect(url).toEqual('http://stream-sight.test:5000');
        });

        it('should return an https URL if TLS is true', () => {
            const config = getStreamSightConfig();
            config.useTLS = true;
            const url = getStreamSightBaseUrl(config);
            expect(url).toEqual('https://stream-sight.test:5000');
        });

        it('should return a URL with the configured port', () => {
            const config = getStreamSightConfig();
            config.rainbowStoragePort = 4711;
            const url = getStreamSightBaseUrl(config);
            expect(url).toEqual('http://stream-sight.test:4711');
        });

    });

});
