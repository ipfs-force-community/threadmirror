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
import type { PaginationMeta } from './PaginationMeta';
import {
    PaginationMetaFromJSON,
    PaginationMetaFromJSONTyped,
    PaginationMetaToJSON,
    PaginationMetaToJSONTyped,
} from './PaginationMeta';
import type { MentionSummary } from './MentionSummary';
import {
    MentionSummaryFromJSON,
    MentionSummaryFromJSONTyped,
    MentionSummaryToJSON,
    MentionSummaryToJSONTyped,
} from './MentionSummary';

/**
 * 
 * @export
 * @interface MentionsGet200Response
 */
export interface MentionsGet200Response {
    /**
     * 
     * @type {Array<MentionSummary>}
     * @memberof MentionsGet200Response
     */
    data?: Array<MentionSummary>;
    /**
     * 
     * @type {PaginationMeta}
     * @memberof MentionsGet200Response
     */
    meta?: PaginationMeta;
}

/**
 * Check if a given object implements the MentionsGet200Response interface.
 */
export function instanceOfMentionsGet200Response(value: object): value is MentionsGet200Response {
    return true;
}

export function MentionsGet200ResponseFromJSON(json: any): MentionsGet200Response {
    return MentionsGet200ResponseFromJSONTyped(json, false);
}

export function MentionsGet200ResponseFromJSONTyped(json: any, ignoreDiscriminator: boolean): MentionsGet200Response {
    if (json == null) {
        return json;
    }
    return {
        
        'data': json['data'] == null ? undefined : ((json['data'] as Array<any>).map(MentionSummaryFromJSON)),
        'meta': json['meta'] == null ? undefined : PaginationMetaFromJSON(json['meta']),
    };
}

export function MentionsGet200ResponseToJSON(json: any): MentionsGet200Response {
    return MentionsGet200ResponseToJSONTyped(json, false);
}

export function MentionsGet200ResponseToJSONTyped(value?: MentionsGet200Response | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'data': value['data'] == null ? undefined : ((value['data'] as Array<any>).map(MentionSummaryToJSON)),
        'meta': PaginationMetaToJSON(value['meta']),
    };
}

