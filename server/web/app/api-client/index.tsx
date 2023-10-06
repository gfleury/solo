'use client'

import useSWR, { SWRResponse } from 'swr'
import useSWRMutation from 'swr/mutation'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL

const OAUTH_BASE_URL = process.env.NEXT_PUBLIC_OAUTH_BASE_URL


function getAuthenticationHeaders(): Headers | undefined {
    return undefined;
}

const fetcher = (input: string) => fetch(input, { credentials: "include", headers: getAuthenticationHeaders() }).then(res => res.json())

async function fetcherPost(url: string, { arg }: any): Promise<Response> {
    return fetch(url, {
        credentials: "include",
        headers: getAuthenticationHeaders(),
        method: 'POST',
        body: JSON.stringify(arg)
    })
}

async function fetcherDelete(url: string, { arg }: any): Promise<Response> {
    return fetch(url.concat(arg), {
        credentials: "include",
        headers: getAuthenticationHeaders(),
        method: 'DELETE',
    })
}

async function fetcherPut(url: string, { arg }: any): Promise<Response> {
    return fetch(url.concat(arg), {
        credentials: "include",
        headers: getAuthenticationHeaders(),
        method: 'PUT',
    })
}

async function fetcherPutWithBody(url: string, { arg }: any): Promise<Response> {
    return fetch(url, {
        credentials: "include",
        headers: getAuthenticationHeaders(),
        method: 'PUT',
        body: JSON.stringify(arg)
    })
}

export function GetUrl(path: string) {
    return API_BASE_URL + path
}

export function Post(path: string) {
    return useSWRMutation(API_BASE_URL + path, fetcherPost)
}

export function Delete(path: string) {
    return useSWRMutation(API_BASE_URL + path, fetcherDelete)
}

export function Put(path: string) {
    return useSWRMutation(API_BASE_URL + path, fetcherPut)
}

export function PutWithBody(path: string) {
    return useSWRMutation(API_BASE_URL + path, fetcherPutWithBody)
}

export function Get(path: string): SWRResponse {
    return useSWR(API_BASE_URL + path, fetcher)
}

export function GetToken(state: string): SWRResponse {
    return useSWR(OAUTH_BASE_URL + "/token?state=" + state, fetcher)
}

export function ForceLogin() {
    return OAUTH_BASE_URL + "/auth/google"
}

export function GetUrlAPI(): string {
    return API_BASE_URL+""
}

export function HandleFailures(r: SWRResponse, router: any): SWRResponse {
    if (r.error) router.push(ForceLogin())
    return r
}