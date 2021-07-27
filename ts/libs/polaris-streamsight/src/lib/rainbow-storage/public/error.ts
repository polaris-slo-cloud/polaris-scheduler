import { IRestResponse } from 'typed-rest-client';

/**
 * Stores information about an error caused by a REST request.
 */
export class RestRequestError extends Error {

    constructor(public response: IRestResponse<any>) {
        super();
    }

}
