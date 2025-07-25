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

import { mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface Image
 */
export interface Image {
    /**
     * Image unique identifier
     * @type {string}
     * @memberof Image
     */
    id?: string | null;
    /**
     * Storage bucket name
     * @type {string}
     * @memberof Image
     */
    bucket: string;
}

/**
 * Check if a given object implements the Image interface.
 */
export function instanceOfImage(value: object): value is Image {
    if (!('bucket' in value) || value['bucket'] === undefined) return false;
    return true;
}

export function ImageFromJSON(json: any): Image {
    return ImageFromJSONTyped(json, false);
}

export function ImageFromJSONTyped(json: any, ignoreDiscriminator: boolean): Image {
    if (json == null) {
        return json;
    }
    return {
        
        'id': json['id'] == null ? undefined : json['id'],
        'bucket': json['bucket'],
    };
}

export function ImageToJSON(json: any): Image {
    return ImageToJSONTyped(json, false);
}

export function ImageToJSONTyped(value?: Image | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'id': value['id'],
        'bucket': value['bucket'],
    };
}

