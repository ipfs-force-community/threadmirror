/* tslint:disable */
/* eslint-disable */
/**
 * Threadmirror Backend API
 * API for managing user profiles and follows
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


import * as runtime from '../runtime';

export interface QrcodeGetRequest {
    threadId: string;
}

/**
 * 
 */
export class QRCodeApi extends runtime.BaseAPI {

    /**
     * Generate and return a QR code image for the given thread id
     * Render QR code
     */
    async qrcodeGetRaw(requestParameters: QrcodeGetRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<Blob>> {
        if (requestParameters['threadId'] == null) {
            throw new runtime.RequiredError(
                'threadId',
                'Required parameter "threadId" was null or undefined when calling qrcodeGet().'
            );
        }

        const queryParameters: any = {};

        if (requestParameters['threadId'] != null) {
            queryParameters['thread_id'] = requestParameters['threadId'];
        }

        const headerParameters: runtime.HTTPHeaders = {};


        let urlPath = `/qrcode`;

        const response = await this.request({
            path: urlPath,
            method: 'GET',
            headers: headerParameters,
            query: queryParameters,
        }, initOverrides);

        return new runtime.BlobApiResponse(response);
    }

    /**
     * Generate and return a QR code image for the given thread id
     * Render QR code
     */
    async qrcodeGet(requestParameters: QrcodeGetRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<Blob> {
        const response = await this.qrcodeGetRaw(requestParameters, initOverrides);
        return await response.value();
    }

}
