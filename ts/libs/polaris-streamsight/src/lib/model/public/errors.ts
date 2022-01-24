import { initSelf } from '@polaris-sloc/core';
import { IRestResponse } from 'typed-rest-client';

/**
 * Stores information about an error caused by a REST request.
 */
export class RestRequestError extends Error {

    url: string;
    response: IRestResponse<any>;
    request: any;
    httpOptions: any;
    cause: any;

    constructor(errorInfo: Partial<RestRequestError>, message?: string) {
        super(message || 'Error executing REST request.');
        initSelf(this, errorInfo);
    }

    toString(): string {
        return JSON.stringify(this, undefined, '  ');
    }

}

/**
 * Stores information about an error caused by a StreamSight request.
 */
export class StreamSightError extends RestRequestError {

    constructor(errorInfo: Partial<StreamSightError>) {
        super(errorInfo, 'Error executing StreamSight request.');
    }

}
