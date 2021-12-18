import { IRestResponse } from 'typed-rest-client';

/**
 * Stores information about an error caused by a REST request.
 */
export class RestRequestError extends Error {

    constructor(public response: IRestResponse<any>, public request?: any, public cause?: any) {
        super('Error executing REST request.');
    }

}

/**
 * Stores information about an error caused by a StreamSight request.
 */
export class StreamSightError extends Error {

    constructor(public response: IRestResponse<any>, public request?: any, public cause?: any) {
        super('Error executing StreamSight request.');
    }

}
